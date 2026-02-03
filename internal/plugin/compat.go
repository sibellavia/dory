package plugin

import "fmt"

// ValidateAPIVersion checks whether a plugin declares a host-compatible API version.
func ValidateAPIVersion(version string) error {
	if version == "" {
		return fmt.Errorf("api_version is required")
	}
	if version != APIVersionV1 {
		return fmt.Errorf("unsupported api_version %q (supported: %s)", version, APIVersionV1)
	}
	return nil
}
