package nomivela

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateNamespaceSendsHeadersAndBody(t *testing.T) {
	var gotIDempotency, gotActor string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/namespaces" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		gotIDempotency = r.Header.Get("Idempotency-Key")
		gotActor = r.Header.Get("X-Actor")

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		if body["namespace"] != "https://auth.example.com" || body["authorityRootRef"] != "root:org-a" {
			t.Errorf("unexpected body: %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Namespace{Namespace: "https://auth.example.com", AuthorityRootRef: "root:org-a", Status: "active", NamespaceEpoch: 1})
	}))
	defer server.Close()

	client := New(server.URL, WithActor("user:ci"))
	namespace, err := client.CreateNamespace(context.Background(), "https://auth.example.com", "root:org-a", Mutation{Reason: "bootstrap"})
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	if namespace.Status != "active" || namespace.NamespaceEpoch != 1 {
		t.Fatalf("unexpected namespace: %+v", namespace)
	}
	if gotIDempotency == "" {
		t.Fatal("Idempotency-Key header missing")
	}
	if gotActor != "user:ci" {
		t.Fatalf("X-Actor = %q, want user:ci", gotActor)
	}
}

func TestListIdentitiesEscapesQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agent-identities" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("namespace"); got != "https://auth.example.com" {
			t.Errorf("namespace = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []AgentIdentity{{AgentID: "abc", State: "bound"}}})
	}))
	defer server.Close()

	identities, err := New(server.URL).ListIdentities(context.Background(), "https://auth.example.com")
	if err != nil {
		t.Fatalf("list identities: %v", err)
	}
	if len(identities) != 1 || identities[0].AgentID != "abc" {
		t.Fatalf("unexpected identities: %+v", identities)
	}
}

func TestAPIErrorCarriesCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "conflict", "message": "an active agent id already exists", "correlationId": "corr-9"},
		})
	}))
	defer server.Close()

	_, err := New(server.URL).CreateNamespace(context.Background(), "https://auth.example.com", "root:org-a", Mutation{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Status != http.StatusConflict || apiErr.Code != "conflict" || apiErr.CorrelationID != "corr-9" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestContainAgentAndCommitInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agents/agent_order_processor/containment":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["mode"] != "quarantine" {
				t.Errorf("mode = %v", body["mode"])
			}
			_ = json.NewEncoder(w).Encode(ContainmentResult{
				AgentRef: "agent_order_processor", Mode: "quarantine", AgentState: "suspended",
				AgentEpoch: 5, Identities: []string{"id-1"}, Instances: []string{"inst-1"},
			})
		case "/v1/agent-identities/id-1/instances":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["leaseExpiresAt"] != "2999-01-01T00:00:00Z" {
				t.Errorf("leaseExpiresAt = %v", body["leaseExpiresAt"])
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(AgentInstance{InstanceID: "inst-1", State: "active", Generation: 1})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := New(server.URL)

	containment, err := client.ContainAgent(context.Background(), "agent_order_processor", "quarantine", Mutation{Reason: "incident"})
	if err != nil {
		t.Fatalf("contain: %v", err)
	}
	if containment.AgentState != "suspended" || len(containment.Instances) != 1 {
		t.Fatalf("unexpected containment: %+v", containment)
	}

	instance, err := client.CommitInstance(context.Background(), "id-1", InstanceCommit{
		Namespace: "https://auth.example.com", WorkloadRegistrationID: "wr-1", WorkloadID: "w",
		ArtifactDigest: "sha256:abc", AttestationRef: "a", LeaseExpiresAt: time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC),
	}, Mutation{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if instance.State != "active" {
		t.Fatalf("unexpected instance: %+v", instance)
	}
}
