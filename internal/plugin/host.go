package plugin

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	healthMethod = "dory.health"
)

// Invoke executes a single plugin request and returns the result payload.
// Plugins are executed as short-lived processes for each invocation.
func Invoke(info PluginInfo, method string, params map[string]interface{}, timeout time.Duration) (map[string]interface{}, string, int64, error) {
	started := time.Now()

	if len(info.Command) == 0 {
		return nil, "", 0, fmt.Errorf("plugin %s has no command in manifest", info.Name)
	}
	if method == "" {
		return nil, "", 0, fmt.Errorf("method is required")
	}
	if err := ValidateAPIVersion(info.APIVersion); err != nil {
		return nil, "", 0, fmt.Errorf("plugin %s: %w", info.Name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	command := info.Command[0]
	if !filepath.IsAbs(command) && strings.Contains(command, string(filepath.Separator)) {
		command = filepath.Join(info.Dir, command)
		if abs, err := filepath.Abs(command); err == nil {
			command = abs
		}
	}

	cmd := exec.CommandContext(ctx, command, info.Command[1:]...)
	if info.Dir != "" {
		if absDir, err := filepath.Abs(info.Dir); err == nil {
			cmd.Dir = absDir
		} else {
			cmd.Dir = info.Dir
		}
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, "", 0, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", 0, err
	}

	if err := cmd.Start(); err != nil {
		return nil, "", 0, err
	}

	req := Request{
		ID:     "req-1",
		Method: method,
		Params: params,
	}

	enc := json.NewEncoder(stdin)
	if err := enc.Encode(req); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return nil, strings.TrimSpace(stderr.String()), 0, err
	}
	_ = stdin.Close()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	if !scanner.Scan() {
		_ = cmd.Wait()
		duration := time.Since(started).Milliseconds()
		if scanner.Err() != nil {
			return nil, strings.TrimSpace(stderr.String()), duration, scanner.Err()
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, strings.TrimSpace(stderr.String()), duration, fmt.Errorf("plugin request timed out")
		}
		if stderr.Len() > 0 {
			return nil, strings.TrimSpace(stderr.String()), duration, fmt.Errorf("plugin returned no response")
		}
		return nil, "", duration, fmt.Errorf("plugin returned no response")
	}

	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		_ = cmd.Wait()
		return nil, strings.TrimSpace(stderr.String()), time.Since(started).Milliseconds(), fmt.Errorf("invalid response: %w", err)
	}

	waitErr := cmd.Wait()
	if waitErr != nil && ctx.Err() == context.DeadlineExceeded {
		return nil, strings.TrimSpace(stderr.String()), time.Since(started).Milliseconds(), fmt.Errorf("plugin request timed out")
	}
	if waitErr != nil && stderr.Len() > 0 {
		return nil, strings.TrimSpace(stderr.String()), time.Since(started).Milliseconds(), fmt.Errorf("plugin process error: %w", waitErr)
	}

	if resp.Error != nil {
		return nil, strings.TrimSpace(stderr.String()), time.Since(started).Milliseconds(), errors.New(resp.Error.Message)
	}

	if resp.Result == nil {
		resp.Result = map[string]interface{}{}
	}

	return resp.Result, strings.TrimSpace(stderr.String()), time.Since(started).Milliseconds(), nil
}

// HealthCheck runs a best-effort health probe against a plugin process.
func HealthCheck(info PluginInfo, timeout time.Duration) HealthStatus {
	status := HealthStatus{
		Name:    info.Name,
		Status:  "error",
		Message: "health probe failed",
	}

	result, stderr, duration, err := Invoke(info, healthMethod, map[string]interface{}{
		"api_version": APIVersionV1,
	}, timeout)
	if err != nil {
		status.Error = err.Error()
		if stderr != "" {
			status.Message = stderr
		}
		status.DurationMS = duration
		return status
	}

	status.Reachable = true
	status.Status = "ok"
	status.Message = "plugin is reachable"
	if raw, ok := result["status"].(string); ok && raw != "" {
		status.Status = raw
	}
	if raw, ok := result["message"].(string); ok && raw != "" {
		status.Message = raw
	}
	if stderr != "" {
		status.Status = "warning"
		status.Message = "plugin health response received with process warning"
		status.Error = stderr
	}

	status.DurationMS = duration
	return status
}
