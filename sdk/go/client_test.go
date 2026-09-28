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

func TestClientSendsBearerToken(t *testing.T) {
	var gotAuth []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, WithToken("secret-token")).ListNamespaces(context.Background()); err != nil {
		t.Fatalf("static token: %v", err)
	}
	if _, err := New(server.URL, WithTokenProvider(func() (string, error) { return "rotated", nil })).ListNamespaces(context.Background()); err != nil {
		t.Fatalf("provider token: %v", err)
	}

	if len(gotAuth) != 2 || gotAuth[0] != "Bearer secret-token" || gotAuth[1] != "Bearer rotated" {
		t.Fatalf("authorization headers = %v", gotAuth)
	}
}

func TestCommitInstanceSendsCallerIdempotencyKey(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		_, _ = w.Write([]byte(`{"instanceId":"i1"}`))
	}))
	defer server.Close()

	_, err := New(server.URL).CommitInstance(context.Background(), "a1", InstanceCommit{
		Namespace: "https://auth.example.com", IdempotencyKey: "enr-1",
		LeaseExpiresAt: time.Now().Add(time.Hour),
	}, Mutation{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if gotKey != "enr-1" {
		t.Fatalf("idempotency key = %q, want enr-1", gotKey)
	}
}

func TestReplayEventsEncodesCursor(t *testing.T) {
	var gotAfter, gotLimit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAfter = r.URL.Query().Get("after")
		gotLimit = r.URL.Query().Get("limit")
		_, _ = w.Write([]byte(`{"items":[{"eventId":"e1","eventType":"agent.active","aggregateType":"agent","aggregateId":"a1","sequence":1,"cursor":5,"payloadVersion":1,"payload":{"objectType":"agent"},"occurredAt":"2026-01-01T00:00:00Z","attempts":0}],"nextCursor":5}`))
	}))
	defer server.Close()

	page, err := New(server.URL).ReplayEvents(context.Background(), 4, 10)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if gotAfter != "4" || gotLimit != "10" {
		t.Fatalf("after=%s limit=%s", gotAfter, gotLimit)
	}
	if len(page.Items) != 1 || page.Items[0].Cursor != 5 || page.Items[0].Payload["objectType"] != "agent" || page.NextCursor != 5 {
		t.Fatalf("page = %+v", page)
	}
}

func TestGetDiscoveryAndJWKS(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path == "/.well-known/agent-iam/jwks.json" {
			_, _ = w.Write([]byte(`{"keys":[{"kty":"OKP","crv":"Ed25519","kid":"k1","alg":"EdDSA","use":"sig","x":"abc"}]}`))
			return
		}
		if r.URL.Query().Get("namespace") != "https://auth.example.com" {
			t.Errorf("namespace = %q", r.URL.Query().Get("namespace"))
		}
		_, _ = w.Write([]byte(`{"discoveryVersion":"1","namespace":"https://auth.example.com","registryEndpoint":"https://registry.example.com","issuer":"https://idp.example.com","jwksUri":"https://idp.example.com/jwks.json","supportedProofProfiles":["private_key_jwt"],"signingKid":"k1","alg":"EdDSA","signature":"sig"}`))
	}))
	defer server.Close()

	document, err := New(server.URL).GetDiscovery(context.Background(), "https://auth.example.com")
	if err != nil {
		t.Fatalf("discovery: %v", err)
	}
	if document.SigningKID != "k1" || document.Signature != "sig" || document.DiscoveryVersion != "1" {
		t.Fatalf("document = %+v", document)
	}
	jwks, err := New(server.URL).GetDiscoveryJWKS(context.Background())
	if err != nil {
		t.Fatalf("jwks: %v", err)
	}
	if len(jwks.Keys) != 1 || jwks.Keys[0].Kid != "k1" {
		t.Fatalf("jwks = %+v", jwks)
	}
}

func TestGetRegistryContextEncodesSelectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/registry-context" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("namespace") != "https://auth.example.com" || r.URL.Query().Get("agentId") != "agent-id" || r.URL.Query().Get("instanceId") != "instance-id" || r.URL.Query().Get("workloadRegistrationId") != "workload-id" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"namespace":{"namespace":"https://auth.example.com","authorityRootRef":"root","status":"active","namespaceEpoch":1},"agent":{"agentRef":"agent_orders","name":"Orders","purpose":"orders","sponsorRef":"org","ownerRef":"owner","riskClass":"low","state":"active","agentEpoch":1},"identity":{"namespace":"https://auth.example.com","agentId":"agent-id","agentRef":"agent_orders","state":"active","identityEpoch":1}}`))
	}))
	defer server.Close()

	context, err := New(server.URL).GetRegistryContext(context.Background(), "https://auth.example.com", "agent-id", "instance-id", "workload-id")
	if err != nil {
		t.Fatalf("get context: %v", err)
	}
	if context.Identity.AgentID != "agent-id" || context.Agent.AgentRef != "agent_orders" {
		t.Fatalf("context = %+v", context)
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
