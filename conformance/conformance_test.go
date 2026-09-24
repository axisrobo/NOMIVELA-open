// Package conformance validates the published contract fixtures against the
// published JSON Schema.
package conformance

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaPath = "../contracts/schemas/agent-registry-v0.1.schema.json"

type manifest struct {
	Cases []struct {
		File       string `json:"file"`
		Definition string `json:"definition"`
		Valid      bool   `json:"valid"`
	} `json:"cases"`
}

func loadManifest(t *testing.T) manifest {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("fixtures", "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m manifest
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(m.Cases) == 0 {
		t.Fatal("manifest has no cases")
	}
	return m
}

func loadSchema(t *testing.T) any {
	t.Helper()
	body, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	return doc
}

func compileDefinition(t *testing.T, schema any, definition string) *jsonschema.Schema {
	t.Helper()
	ref, err := json.Marshal(map[string]any{"$ref": "registry.json#/$defs/" + definition})
	if err != nil {
		t.Fatalf("marshal ref: %v", err)
	}
	refDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(ref))
	if err != nil {
		t.Fatalf("decode ref: %v", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("registry.json", schema); err != nil {
		t.Fatalf("add schema: %v", err)
	}
	if err := compiler.AddResource("check.json", refDoc); err != nil {
		t.Fatalf("add ref: %v", err)
	}
	compiled, err := compiler.Compile("check.json")
	if err != nil {
		t.Fatalf("compile %s: %v", definition, err)
	}
	return compiled
}

func TestContractFixtures(t *testing.T) {
	m := loadManifest(t)
	schema := loadSchema(t)

	for _, tc := range m.Cases {
		t.Run(tc.File, func(t *testing.T) {
			compiled := compileDefinition(t, schema, tc.Definition)

			body, err := os.ReadFile(filepath.Join("fixtures", tc.File))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(body))
			if err != nil {
				t.Fatalf("decode fixture: %v", err)
			}

			err = compiled.Validate(instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid %s fixture, got: %v", tc.Definition, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid %s fixture, but schema accepted it", tc.Definition)
			}
		})
	}
}
