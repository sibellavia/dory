package fileio

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReadYAML reads a YAML file into the provided struct
func ReadYAML(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// WriteYAML writes a struct to a YAML file
func WriteYAML(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// EngFile represents a parsed .eng file with YAML frontmatter and Markdown body
type EngFile struct {
	Frontmatter map[string]interface{}
	Body        string
}

// ParseEngFile parses a .eng file with YAML frontmatter and Markdown body
func ParseEngFile(path string) (*EngFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Check for YAML frontmatter delimiters
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("file does not start with YAML frontmatter")
	}

	// Find the closing delimiter
	rest := content[4:] // Skip opening "---\n"
	endIdx := strings.Index(rest, "\n---\n")
	if endIdx == -1 {
		// Try with just "---" at end of frontmatter
		endIdx = strings.Index(rest, "\n---")
		if endIdx == -1 {
			return nil, fmt.Errorf("no closing YAML frontmatter delimiter found")
		}
	}

	frontmatterStr := rest[:endIdx]
	bodyStart := endIdx + 5 // Skip "\n---\n"
	if bodyStart > len(rest) {
		bodyStart = len(rest)
	}
	body := ""
	if bodyStart < len(rest) {
		body = strings.TrimPrefix(rest[bodyStart:], "\n")
	}

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &EngFile{
		Frontmatter: frontmatter,
		Body:        body,
	}, nil
}

// WriteEngFile writes an .eng file with YAML frontmatter and Markdown body
func WriteEngFile(path string, frontmatter interface{}, body string) error {
	var buf bytes.Buffer

	buf.WriteString("---\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(frontmatter); err != nil {
		return err
	}
	enc.Close()

	buf.WriteString("---\n\n")
	buf.WriteString(body)

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// NextID generates the next ID for a given prefix (L, D, P)
// Scans the directory for existing files and returns the next sequential ID
func NextID(dir string, prefix string) (string, error) {
	pattern := filepath.Join(dir, prefix+"*.eng")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	maxNum := 0
	re := regexp.MustCompile(fmt.Sprintf(`%s(\d+)\.eng$`, prefix))

	for _, match := range matches {
		base := filepath.Base(match)
		if m := re.FindStringSubmatch(base); m != nil {
			if num, err := strconv.Atoi(m[1]); err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	return fmt.Sprintf("%s%03d", prefix, maxNum+1), nil
}

// ListEngFiles returns all .eng files in a directory
func ListEngFiles(dir string) ([]string, error) {
	pattern := filepath.Join(dir, "*.eng")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadFileContent reads the entire content of a file as a string
func ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetEditor returns the user's preferred editor from environment
func GetEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	return "vi" // fallback
}

// ReadLines reads a file line by line
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
