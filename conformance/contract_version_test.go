package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestContractVersionIsStable guards the public contract version. The release
// version advances independently; the contract version changes only through an
// approved compatibility decision, which must update this test and
// COMPATIBILITY.md together.
func TestContractVersionIsStable(t *testing.T) {
	expected := []string{
		"../contracts/agent-registry-v0.1.openapi.yaml",
		"../contracts/schemas/agent-registry-v0.1.schema.json",
		"../contracts/schemas/eidovela-v1-compat.schema.json",
	}
	for _, path := range expected {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected contract file %s: %v", filepath.Base(path), err)
		}
	}

	openapi, err := os.ReadFile("../contracts/agent-registry-v0.1.openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	if !strings.Contains(string(openapi), "version: 0.1.0") {
		t.Fatal("agent-registry contract version must be 0.1.0; see COMPATIBILITY.md before changing it")
	}

	assertSchemaID(t, "../contracts/schemas/agent-registry-v0.1.schema.json", "agent-registry-v0.1")
	assertSchemaID(t, "../contracts/schemas/eidovela-v1-compat.schema.json", "eidovela-v1-compat")
}

func assertSchemaID(t *testing.T, path, wantFragment string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var schema struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(body, &schema); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if !strings.Contains(schema.ID, wantFragment) {
		t.Fatalf("%s $id = %q, want it to contain %q", filepath.Base(path), schema.ID, wantFragment)
	}
}
