package plugin

import "testing"

func TestValidateAPIVersion(t *testing.T) {
	t.Run("supported", func(t *testing.T) {
		if err := ValidateAPIVersion(APIVersionV1); err != nil {
			t.Fatalf("expected supported version, got err: %v", err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		if err := ValidateAPIVersion(""); err == nil {
			t.Fatal("expected missing version error")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		if err := ValidateAPIVersion("v2"); err == nil {
			t.Fatal("expected unsupported version error")
		}
	})
}
