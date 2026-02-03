package plugin

import (
	"strings"
	"testing"
	"time"
)

func TestValidateCustomTypeValid(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","result":{"valid":true,"message":"ok"}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	validation, err := ValidateCustomType(info, "incident", "DB outage", "ops", "body", []string{"L001"}, 2*time.Second)
	if err != nil {
		t.Fatalf("validate custom type: %v", err)
	}
	if !validation.Valid {
		t.Fatalf("expected valid result, got %+v", validation)
	}
	if validation.Message != "ok" {
		t.Fatalf("unexpected message: %+v", validation)
	}
}

func TestValidateCustomTypeRejected(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","result":{"valid":false,"message":"schema mismatch","errors":["field x is required"]}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	validation, err := ValidateCustomType(info, "incident", "DB outage", "ops", "body", nil, 2*time.Second)
	if err == nil {
		t.Fatal("expected validation rejection error")
	}
	if !strings.Contains(err.Error(), "schema mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
	if validation == nil || validation.Valid {
		t.Fatalf("expected invalid validation result, got %+v", validation)
	}
	if len(validation.Errors) != 1 || validation.Errors[0] != "field x is required" {
		t.Fatalf("unexpected validation errors: %+v", validation)
	}
}

func TestValidateCustomTypeMissingValidField(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","result":{"message":"missing valid"}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	_, err := ValidateCustomType(info, "incident", "DB outage", "ops", "body", nil, 2*time.Second)
	if err == nil {
		t.Fatal("expected invalid response error")
	}
	if !strings.Contains(err.Error(), "missing boolean field \"valid\"") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCustomTypePluginError(t *testing.T) {
	script := makeScript(t, `read line
echo '{"id":"req-1","error":{"code":500,"message":"not implemented"}}'`)
	info := PluginInfo{Name: "demo", APIVersion: APIVersionV1, Command: []string{script}}

	_, err := ValidateCustomType(info, "incident", "DB outage", "ops", "body", nil, 2*time.Second)
	if err == nil {
		t.Fatal("expected plugin error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
