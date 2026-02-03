package plugin

import (
	"fmt"
	"strings"
	"time"
)

const typeValidateMethod = "dory.type.validate"

// TypeValidation captures plugin-provided custom type validation output.
type TypeValidation struct {
	Valid      bool     `json:"valid" yaml:"valid"`
	Message    string   `json:"message,omitempty" yaml:"message,omitempty"`
	Errors     []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	DurationMS int64    `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
	Stderr     string   `json:"stderr,omitempty" yaml:"stderr,omitempty"`
}

// ValidateCustomType asks a plugin to validate a custom type payload.
// Plugins must return {"valid": <bool>} and may also return message/errors.
func ValidateCustomType(info PluginInfo, typeName, oneliner, topic, body string, refs []string, timeout time.Duration) (*TypeValidation, error) {
	result, stderr, durationMS, err := Invoke(info, typeValidateMethod, map[string]interface{}{
		"api_version": APIVersionV1,
		"type":        typeName,
		"oneliner":    oneliner,
		"topic":       topic,
		"body":        body,
		"refs":        refs,
	}, timeout)
	if err != nil {
		return &TypeValidation{
			Valid:      false,
			DurationMS: durationMS,
			Stderr:     stderr,
		}, fmt.Errorf("custom type %q validation failed via plugin %q: %w", typeName, info.Name, err)
	}

	validation := &TypeValidation{
		DurationMS: durationMS,
		Stderr:     stderr,
	}

	if message, ok := result["message"].(string); ok {
		validation.Message = message
	}
	validation.Errors = parseValidationErrors(result["errors"])

	valid, ok := result["valid"].(bool)
	if !ok {
		return validation, fmt.Errorf("plugin %q returned invalid validation response for type %q: missing boolean field \"valid\"", info.Name, typeName)
	}
	validation.Valid = valid
	if valid {
		return validation, nil
	}

	reason := validation.Message
	if reason == "" && len(validation.Errors) > 0 {
		reason = strings.Join(validation.Errors, "; ")
	}
	if reason == "" {
		reason = "validation rejected by plugin"
	}
	return validation, fmt.Errorf("custom type %q rejected by plugin %q: %s", typeName, info.Name, reason)
}

func parseValidationErrors(raw interface{}) []string {
	values, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	errors := make([]string, 0, len(values))
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			errors = append(errors, s)
		}
	}
	return errors
}
