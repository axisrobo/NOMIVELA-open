// Package conformance validates the published contract fixtures against the
// published JSON Schemas.
package conformance

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type manifest struct {
	DefaultSchema string `json:"defaultSchema"`
	Cases         []struct {
		File       string `json:"file"`
		Schema     string `json:"schema"`
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
	if m.DefaultSchema == "" {
		t.Fatal("manifest is missing defaultSchema")
	}
	return m
}

func loadJSON(t *testing.T, path string) any {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return doc
}

func compileDefinition(t *testing.T, schemaPath, definition string) *jsonschema.Schema {
	t.Helper()
	schema := loadJSON(t, schemaPath)
	resource := filepath.Base(schemaPath)

	ref, err := json.Marshal(map[string]any{"$ref": resource + "#/$defs/" + definition})
	if err != nil {
		t.Fatalf("marshal ref: %v", err)
	}
	refDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(ref))
	if err != nil {
		t.Fatalf("decode ref: %v", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(resource, schema); err != nil {
		t.Fatalf("add schema %s: %v", schemaPath, err)
	}
	if err := compiler.AddResource("check-"+resource, refDoc); err != nil {
		t.Fatalf("add ref: %v", err)
	}
	compiled, err := compiler.Compile("check-" + resource)
	if err != nil {
		t.Fatalf("compile %s#%s: %v", schemaPath, definition, err)
	}
	return compiled
}

func TestContractFixtures(t *testing.T) {
	m := loadManifest(t)

	for _, tc := range m.Cases {
		t.Run(tc.File, func(t *testing.T) {
			schemaPath := tc.Schema
			if schemaPath == "" {
				schemaPath = m.DefaultSchema
			}
			compiled := compileDefinition(t, schemaPath, tc.Definition)

			instance := loadJSON(t, filepath.Join("fixtures", tc.File))

			err := compiled.Validate(instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid %s fixture, got: %v", tc.Definition, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid %s fixture, but schema accepted it", tc.Definition)
			}
		})
	}
}
