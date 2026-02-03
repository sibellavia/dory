package plugin

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const hookRunMethod = "dory.hook.run"

// HookEvent is a lifecycle event exposed to plugins.
type HookEvent string

const (
	HookBeforeCreate HookEvent = "before_create"
	HookAfterCreate  HookEvent = "after_create"
	HookBeforeRemove HookEvent = "before_remove"
	HookAfterRemove  HookEvent = "after_remove"
	HookAfterCompact HookEvent = "after_compact"
)

var hookEventSet = map[HookEvent]struct{}{
	HookBeforeCreate: {},
	HookAfterCreate:  {},
	HookBeforeRemove: {},
	HookAfterRemove:  {},
	HookAfterCompact: {},
}

// HookResult is a single plugin hook invocation result.
type HookResult struct {
	Plugin     string `json:"plugin" yaml:"plugin"`
	Event      string `json:"event" yaml:"event"`
	Status     string `json:"status" yaml:"status"` // ok, warning, error, blocked
	Message    string `json:"message,omitempty" yaml:"message,omitempty"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
}

// RunHooks executes enabled plugins that expose the given hook event.
// Hook errors are reported in results but are fail-soft by default.
// For before_* hooks, plugins can block by returning {"allow": false}.
func RunHooks(doryRoot string, event HookEvent, context map[string]interface{}, timeout time.Duration) ([]HookResult, error) {
	if _, ok := hookEventSet[event]; !ok {
		return nil, fmt.Errorf("unsupported hook event %q", event)
	}

	plugins, _, err := Discover(doryRoot)
	if err != nil {
		return nil, err
	}

	targets := make([]PluginInfo, 0)
	for _, p := range plugins {
		if !p.Enabled {
			continue
		}
		for _, hook := range p.Capabilities.Hooks {
			if hook == string(event) {
				targets = append(targets, p)
				break
			}
		}
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Name < targets[j].Name
	})

	results := make([]HookResult, 0, len(targets))
	blockedBy := make([]string, 0)

	for _, p := range targets {
		result, stderr, durationMS, invokeErr := Invoke(p, hookRunMethod, map[string]interface{}{
			"api_version": APIVersionV1,
			"event":       string(event),
			"context":     context,
		}, timeout)
		if invokeErr != nil {
			results = append(results, HookResult{
				Plugin:     p.Name,
				Event:      string(event),
				Status:     "error",
				Message:    "hook invocation failed",
				Error:      invokeErr.Error(),
				DurationMS: durationMS,
			})
			continue
		}

		hookResult := HookResult{
			Plugin:     p.Name,
			Event:      string(event),
			Status:     "ok",
			DurationMS: durationMS,
		}
		if msg, ok := result["message"].(string); ok && msg != "" {
			hookResult.Message = msg
		}
		if stderr != "" {
			hookResult.Status = "warning"
			hookResult.Error = stderr
			if hookResult.Message == "" {
				hookResult.Message = "hook completed with plugin stderr"
			}
		}

		if strings.HasPrefix(string(event), "before_") {
			if allow, ok := result["allow"].(bool); ok && !allow {
				hookResult.Status = "blocked"
				if hookResult.Message == "" {
					hookResult.Message = "operation blocked by plugin hook"
				}
				blockedBy = append(blockedBy, p.Name)
			}
		}

		results = append(results, hookResult)
	}

	if len(blockedBy) > 0 {
		return results, fmt.Errorf("operation blocked by hook(s): %s", strings.Join(blockedBy, ", "))
	}
	return results, nil
}
