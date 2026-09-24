package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T) *bytes.Buffer {
	t.Helper()
	buffer := &bytes.Buffer{}
	original := stdout
	stdout = buffer
	t.Cleanup(func() { stdout = original })
	return buffer
}

func TestNamespaceCreateCommand(t *testing.T) {
	var gotActor, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotActor = r.Header.Get("X-Actor")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"namespace": "https://auth.example.com", "authorityRootRef": "root:org-a", "status": "active", "namespaceEpoch": 1,
		})
	}))
	defer server.Close()

	out := captureStdout(t)
	err := run(context.Background(), []string{
		"namespace", "create",
		"--api", server.URL,
		"--actor", "user:cli",
		"--namespace", "https://auth.example.com",
		"--root-ref", "root:org-a",
		"--reason", "bootstrap",
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if gotPath != "/v1/namespaces" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotActor != "user:cli" {
		t.Fatalf("actor = %q", gotActor)
	}
	if !strings.Contains(out.String(), `"status": "active"`) {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestUnknownCommandReturnsError(t *testing.T) {
	if err := run(context.Background(), []string{"frobnicate"}); err == nil {
		t.Fatal("expected error for unknown command")
	}
	if err := run(context.Background(), nil); err == nil {
		t.Fatal("expected usage error for empty arguments")
	}
}

func TestWorkloadCreateCommandParsesSelector(t *testing.T) {
	var gotSelector map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Selector map[string]string `json:"selector"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotSelector = body.Selector
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workloadRegistrationId": "wr-1", "namespace": "https://auth.example.com",
			"platform": "kubernetes", "selector": body.Selector, "trustDomain": "cluster.local",
			"allowedProofMethods": []string{"spiffe"}, "status": "active", "workloadEpoch": 1,
		})
	}))
	defer server.Close()

	_ = captureStdout(t)
	err := run(context.Background(), []string{
		"workload", "create",
		"--api", server.URL,
		"--namespace", "https://auth.example.com",
		"--platform", "kubernetes",
		"--selector", "app=orders,tier=backend",
		"--trust-domain", "cluster.local",
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if gotSelector["app"] != "orders" || gotSelector["tier"] != "backend" {
		t.Fatalf("selector = %v", gotSelector)
	}
}

func TestParseSelectorRejectsMalformed(t *testing.T) {
	if _, err := parseSelector("apporders"); err == nil {
		t.Fatal("expected error for malformed selector")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
