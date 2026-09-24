// Package nomivela is a minimal Go client for the NOMIVELA Agent Registry API.
//
// It uses only the standard library. Every mutating call sends an
// Idempotency-Key and an X-Actor header, as the contract requires.
package nomivela

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a NOMIVELA Agent Registry API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	actor      string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithActor sets the X-Actor value used for mutations.
func WithActor(actor string) Option {
	return func(c *Client) { c.actor = actor }
}

// New returns a Client for the given base URL.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		actor:      "nomivela-sdk",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Mutation carries the attribution recorded for a state change.
type Mutation struct {
	Reason      string `json:"reason,omitempty"`
	EvidenceRef string `json:"evidenceRef,omitempty"`
}

// Namespace is an authority namespace.
type Namespace struct {
	Namespace        string `json:"namespace"`
	Parent           string `json:"parent,omitempty"`
	AuthorityRootRef string `json:"authorityRootRef"`
	Status           string `json:"status"`
	NamespaceEpoch   int64  `json:"namespaceEpoch"`
}

// Agent is a business Agent Record.
type Agent struct {
	AgentRef   string `json:"agentRef"`
	Name       string `json:"name"`
	Purpose    string `json:"purpose"`
	SponsorRef string `json:"sponsorRef"`
	OwnerRef   string `json:"ownerRef"`
	RiskClass  string `json:"riskClass"`
	State      string `json:"state"`
	AgentEpoch int64  `json:"agentEpoch"`
}

// AgentIdentity is a security Agent Identity Record.
type AgentIdentity struct {
	Namespace         string `json:"namespace"`
	AgentID           string `json:"agentId"`
	AgentRef          string `json:"agentRef"`
	State             string `json:"state"`
	IdentityEpoch     int64  `json:"identityEpoch"`
	AuthorityRootRef  string `json:"authorityRootRef,omitempty"`
	AuthorityRootType string `json:"authorityRootType,omitempty"`
}

// WorkloadRegistration is an approved deployment declaration.
type WorkloadRegistration struct {
	WorkloadRegistrationID string            `json:"workloadRegistrationId"`
	Namespace              string            `json:"namespace"`
	Platform               string            `json:"platform"`
	Selector               map[string]string `json:"selector"`
	TrustDomain            string            `json:"trustDomain"`
	AllowedProofMethods    []string          `json:"allowedProofMethods"`
	Status                 string            `json:"status"`
	WorkloadEpoch          int64             `json:"workloadEpoch"`
}

// AgentInstance is a running instance of an Agent identity.
type AgentInstance struct {
	InstanceID             string `json:"instanceId"`
	Namespace              string `json:"namespace"`
	AgentID                string `json:"agentId"`
	WorkloadRegistrationID string `json:"workloadRegistrationId"`
	WorkloadID             string `json:"workloadId"`
	ArtifactDigest         string `json:"artifactDigest"`
	AttestationRef         string `json:"attestationRef"`
	LeaseExpiresAt         string `json:"leaseExpiresAt"`
	State                  string `json:"state"`
	Generation             int64  `json:"generation"`
}

// LifecycleEvent is append-only evidence of a state transition.
type LifecycleEvent struct {
	EventID       string `json:"eventId"`
	ObjectType    string `json:"objectType"`
	ObjectID      string `json:"objectId"`
	PreviousState string `json:"previousState,omitempty"`
	NewState      string `json:"newState"`
	EpochKind     string `json:"epochKind"`
	Epoch         int64  `json:"epoch"`
	Actor         string `json:"actor"`
	Reason        string `json:"reason,omitempty"`
	EvidenceRef   string `json:"evidenceRef,omitempty"`
	OccurredAt    string `json:"occurredAt"`
}

// OutboxEvent is a transactional outbox record.
type OutboxEvent struct {
	EventID       string `json:"eventId"`
	EventType     string `json:"eventType"`
	AggregateType string `json:"aggregateType"`
	AggregateID   string `json:"aggregateId"`
	Sequence      int64  `json:"sequence"`
	OccurredAt    string `json:"occurredAt"`
}

// ContainmentResult summarizes a cross-namespace containment operation.
type ContainmentResult struct {
	AgentRef   string   `json:"agentRef"`
	Mode       string   `json:"mode"`
	AgentState string   `json:"agentState"`
	AgentEpoch int64    `json:"agentEpoch"`
	Identities []string `json:"identities"`
	Instances  []string `json:"instances"`
}

// APIError is an error response from the API.
type APIError struct {
	Status        int
	Code          string
	Message       string
	CorrelationID string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("nomivela: %s (status %d, code %s, correlation %s)", e.Message, e.Status, e.Code, e.CorrelationID)
}

type listResponse[T any] struct {
	Items []T `json:"items"`
}

// CreateNamespace creates an authority namespace.
func (c *Client) CreateNamespace(ctx context.Context, namespace, authorityRootRef string, mut Mutation) (*Namespace, error) {
	body := struct {
		Mutation
		Namespace        string `json:"namespace"`
		AuthorityRootRef string `json:"authorityRootRef"`
	}{mut, namespace, authorityRootRef}
	var out Namespace
	if err := c.do(ctx, http.MethodPost, "/v1/namespaces", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListNamespaces returns all authority namespaces.
func (c *Client) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	var out listResponse[Namespace]
	if err := c.do(ctx, http.MethodGet, "/v1/namespaces", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// CreateAgent registers an Agent Record.
func (c *Client) CreateAgent(ctx context.Context, agent Agent, mut Mutation) (*Agent, error) {
	body := struct {
		Mutation
		Agent
	}{mut, agent}
	var out Agent
	if err := c.do(ctx, http.MethodPost, "/v1/agents", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAgents returns all Agent Records.
func (c *Client) ListAgents(ctx context.Context) ([]Agent, error) {
	var out listResponse[Agent]
	if err := c.do(ctx, http.MethodGet, "/v1/agents", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// TransitionAgent advances an Agent lifecycle.
func (c *Client) TransitionAgent(ctx context.Context, agentRef, state string, mut Mutation) (*Agent, error) {
	body := struct {
		Mutation
		State string `json:"state"`
	}{mut, state}
	var out Agent
	if err := c.do(ctx, http.MethodPost, "/v1/agents/"+agentRef+"/lifecycle", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ContainAgent contains an Agent across all namespaces.
func (c *Client) ContainAgent(ctx context.Context, agentRef, mode string, mut Mutation) (*ContainmentResult, error) {
	body := struct {
		Mutation
		Mode string `json:"mode"`
	}{mut, mode}
	var out ContainmentResult
	if err := c.do(ctx, http.MethodPost, "/v1/agents/"+agentRef+"/containment", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AllocateIdentity allocates an Agent ID in a namespace.
func (c *Client) AllocateIdentity(ctx context.Context, namespace, agentRef string, mut Mutation) (*AgentIdentity, error) {
	body := struct {
		Mutation
		Namespace string `json:"namespace"`
		AgentRef  string `json:"agentRef"`
	}{mut, namespace, agentRef}
	var out AgentIdentity
	if err := c.do(ctx, http.MethodPost, "/v1/agent-identities", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListIdentities returns the Agent identities of a namespace.
func (c *Client) ListIdentities(ctx context.Context, namespace string) ([]AgentIdentity, error) {
	var out listResponse[AgentIdentity]
	if err := c.do(ctx, http.MethodGet, "/v1/agent-identities?namespace="+urlQueryEscape(namespace), nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// BindIdentity attaches the immutable Authority Binding.
func (c *Client) BindIdentity(ctx context.Context, namespace, agentID, rootRef, rootType string, mut Mutation) (*AgentIdentity, error) {
	body := struct {
		Mutation
		Namespace         string `json:"namespace"`
		AuthorityRootRef  string `json:"authorityRootRef"`
		AuthorityRootType string `json:"authorityRootType"`
	}{mut, namespace, rootRef, rootType}
	var out AgentIdentity
	if err := c.do(ctx, http.MethodPost, "/v1/agent-identities/"+agentID+"/binding", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TransitionIdentity advances an Agent identity lifecycle.
func (c *Client) TransitionIdentity(ctx context.Context, namespace, agentID, state string, mut Mutation) (*AgentIdentity, error) {
	body := struct {
		Mutation
		Namespace string `json:"namespace"`
		State     string `json:"state"`
	}{mut, namespace, state}
	var out AgentIdentity
	if err := c.do(ctx, http.MethodPost, "/v1/agent-identities/"+agentID+"/lifecycle", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateWorkloadRegistration registers an approved workload.
func (c *Client) CreateWorkloadRegistration(ctx context.Context, workload WorkloadRegistration, mut Mutation) (*WorkloadRegistration, error) {
	body := struct {
		Mutation
		Namespace           string            `json:"namespace"`
		Platform            string            `json:"platform"`
		Selector            map[string]string `json:"selector"`
		TrustDomain         string            `json:"trustDomain"`
		AllowedProofMethods []string          `json:"allowedProofMethods"`
	}{mut, workload.Namespace, workload.Platform, workload.Selector, workload.TrustDomain, workload.AllowedProofMethods}
	var out WorkloadRegistration
	if err := c.do(ctx, http.MethodPost, "/v1/workload-registrations", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWorkloadRegistrations returns the workload registrations of a namespace.
func (c *Client) ListWorkloadRegistrations(ctx context.Context, namespace string) ([]WorkloadRegistration, error) {
	var out listResponse[WorkloadRegistration]
	if err := c.do(ctx, http.MethodGet, "/v1/workload-registrations?namespace="+urlQueryEscape(namespace), nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// TransitionWorkloadRegistration advances a workload lifecycle.
func (c *Client) TransitionWorkloadRegistration(ctx context.Context, workloadRegistrationID, state string, mut Mutation) (*WorkloadRegistration, error) {
	body := struct {
		Mutation
		State string `json:"state"`
	}{mut, state}
	var out WorkloadRegistration
	if err := c.do(ctx, http.MethodPost, "/v1/workload-registrations/"+workloadRegistrationID+"/lifecycle", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// InstanceCommit is the verified enrollment submitted for an Agent instance.
type InstanceCommit struct {
	Namespace              string    `json:"namespace"`
	WorkloadRegistrationID string    `json:"workloadRegistrationId"`
	WorkloadID             string    `json:"workloadId"`
	ArtifactDigest         string    `json:"artifactDigest"`
	AttestationRef         string    `json:"attestationRef"`
	LeaseExpiresAt         time.Time `json:"-"`
}

// CommitInstance records a verified enrollment and activates an Agent instance.
func (c *Client) CommitInstance(ctx context.Context, agentID string, commit InstanceCommit, mut Mutation) (*AgentInstance, error) {
	body := struct {
		Mutation
		Namespace              string `json:"namespace"`
		WorkloadRegistrationID string `json:"workloadRegistrationId"`
		WorkloadID             string `json:"workloadId"`
		ArtifactDigest         string `json:"artifactDigest"`
		AttestationRef         string `json:"attestationRef"`
		LeaseExpiresAt         string `json:"leaseExpiresAt"`
	}{
		Mutation:               mut,
		Namespace:              commit.Namespace,
		WorkloadRegistrationID: commit.WorkloadRegistrationID,
		WorkloadID:             commit.WorkloadID,
		ArtifactDigest:         commit.ArtifactDigest,
		AttestationRef:         commit.AttestationRef,
		LeaseExpiresAt:         commit.LeaseExpiresAt.UTC().Format(time.RFC3339),
	}
	var out AgentInstance
	if err := c.do(ctx, http.MethodPost, "/v1/agent-identities/"+agentID+"/instances", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListInstances returns the instances of an Agent identity.
func (c *Client) ListInstances(ctx context.Context, namespace, agentID string) ([]AgentInstance, error) {
	var out listResponse[AgentInstance]
	if err := c.do(ctx, http.MethodGet, "/v1/agent-identities/"+agentID+"/instances?namespace="+urlQueryEscape(namespace), nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// TransitionInstance advances an Agent instance lifecycle.
func (c *Client) TransitionInstance(ctx context.Context, instanceID, state string, mut Mutation) (*AgentInstance, error) {
	body := struct {
		Mutation
		State string `json:"state"`
	}{mut, state}
	var out AgentInstance
	if err := c.do(ctx, http.MethodPost, "/v1/instances/"+instanceID+"/lifecycle", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEvidence returns the append-only lifecycle evidence.
func (c *Client) ListEvidence(ctx context.Context) ([]LifecycleEvent, error) {
	var out listResponse[LifecycleEvent]
	if err := c.do(ctx, http.MethodGet, "/v1/evidence", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// ListEvents returns the transactional outbox events.
func (c *Client) ListEvents(ctx context.Context) ([]OutboxEvent, error) {
	var out listResponse[OutboxEvent]
	if err := c.do(ctx, http.MethodGet, "/v1/events", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("nomivela: encode request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("nomivela: build request: %w", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		request.Header.Set("Idempotency-Key", newIdempotencyKey())
		request.Header.Set("X-Actor", c.actor)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("nomivela: request failed: %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("nomivela: read response: %w", err)
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{Status: response.StatusCode, Message: response.Status}
		var envelope struct {
			Error struct {
				Code          string `json:"code"`
				Message       string `json:"message"`
				CorrelationID string `json:"correlationId"`
			} `json:"error"`
		}
		if json.Unmarshal(payload, &envelope) == nil && envelope.Error.Code != "" {
			apiErr.Code = envelope.Error.Code
			apiErr.Message = envelope.Error.Message
			apiErr.CorrelationID = envelope.Error.CorrelationID
		}
		return apiErr
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("nomivela: decode response: %w", err)
	}
	return nil
}

func newIdempotencyKey() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func urlQueryEscape(value string) string {
	return url.QueryEscape(value)
}
