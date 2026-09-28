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
	"strconv"
	"strings"
	"time"
)

// Client is a NOMIVELA Agent Registry API client.
type Client struct {
	baseURL       string
	httpClient    *http.Client
	actor         string
	tokenProvider func() (string, error)
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

// WithToken sets a static bearer token sent on every request. Use it when the
// deployment enables API tokens or scoped service principals.
func WithToken(token string) Option {
	return func(c *Client) {
		if token == "" {
			c.tokenProvider = nil
			return
		}
		c.tokenProvider = func() (string, error) { return token, nil }
	}
}

// WithTokenProvider sets a bearer token source resolved on every request, so a
// rotated or refreshed token is picked up without rebuilding the client.
func WithTokenProvider(provider func() (string, error)) Option {
	return func(c *Client) { c.tokenProvider = provider }
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
//
// ExpectedEpoch is an optional optimistic-concurrency precondition: when
// non-zero the object's current epoch must match or the mutation returns a
// conflict.
type Mutation struct {
	Reason        string `json:"reason,omitempty"`
	EvidenceRef   string `json:"evidenceRef,omitempty"`
	ExpectedEpoch int64  `json:"expectedEpoch,omitempty"`
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
	AgentRef    string   `json:"agentRef"`
	Name        string   `json:"name"`
	Purpose     string   `json:"purpose"`
	SponsorRef  string   `json:"sponsorRef"`
	OwnerRef    string   `json:"ownerRef"`
	RiskClass   string   `json:"riskClass"`
	AgentClass  string   `json:"agentClass,omitempty"`
	CarrierRefs []string `json:"carrierRefs,omitempty"`
	State       string   `json:"state"`
	AgentEpoch  int64    `json:"agentEpoch"`
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

// ProofMethod is one acceptable attestation method with its profile version.
type ProofMethod struct {
	Method         string `json:"method"`
	ProfileVersion string `json:"profileVersion"`
}

// ProofRequirements is the versioned proof profile for a workload.
type ProofRequirements struct {
	SchemaVersion             string        `json:"schemaVersion"`
	Methods                   []ProofMethod `json:"methods"`
	ExpectedIssuer            string        `json:"expectedIssuer,omitempty"`
	ExpectedAudience          string        `json:"expectedAudience,omitempty"`
	TrustDomain               string        `json:"trustDomain,omitempty"`
	SelectorSchemaVersion     string        `json:"selectorSchemaVersion"`
	AttestationDigestRequired bool          `json:"attestationDigestRequired,omitempty"`
	VerifierIdentity          string        `json:"verifierIdentity,omitempty"`
}

// WorkloadRegistration is an approved deployment declaration.
type WorkloadRegistration struct {
	WorkloadRegistrationID string             `json:"workloadRegistrationId"`
	Namespace              string             `json:"namespace"`
	Platform               string             `json:"platform"`
	Selector               map[string]string  `json:"selector"`
	TrustDomain            string             `json:"trustDomain"`
	AllowedProofMethods    []string           `json:"allowedProofMethods"`
	ProofRequirements      *ProofRequirements `json:"proofRequirements,omitempty"`
	Status                 string             `json:"status"`
	WorkloadEpoch          int64              `json:"workloadEpoch"`
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

// RegistryContext is an atomic Registry snapshot for one identity decision.
// Namespace, Agent, and Identity are always present. WorkloadRegistration and
// Instance are included when selected, or when an instance selects its workload.
type RegistryContext struct {
	Namespace            Namespace             `json:"namespace"`
	Agent                Agent                 `json:"agent"`
	Identity             AgentIdentity         `json:"identity"`
	WorkloadRegistration *WorkloadRegistration `json:"workloadRegistration,omitempty"`
	Instance             *AgentInstance        `json:"instance,omitempty"`
}

// DiscoveryDocument is the registry-published discovery metadata. A signed
// deployment adds DiscoveryVersion, IssuedAt, ExpiresAt, SigningKID, Alg, and
// Signature; a consumer verifies Signature with the key named by SigningKID.
type DiscoveryDocument struct {
	DiscoveryVersion       string   `json:"discoveryVersion,omitempty"`
	Namespace              string   `json:"namespace"`
	RegistryEndpoint       string   `json:"registryEndpoint"`
	Issuer                 string   `json:"issuer"`
	JWKSUri                string   `json:"jwksUri"`
	SupportedProofProfiles []string `json:"supportedProofProfiles"`
	SupportedArtifactTypes []string `json:"supportedArtifactTypes,omitempty"`
	KeyRotation            string   `json:"keyRotation,omitempty"`
	IssuedAt               string   `json:"issuedAt,omitempty"`
	ExpiresAt              string   `json:"expiresAt,omitempty"`
	SigningKID             string   `json:"signingKid,omitempty"`
	Alg                    string   `json:"alg,omitempty"`
	Signature              string   `json:"signature,omitempty"`
}

// JWK is a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	X   string `json:"x"`
}

// JWKS is a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
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

// OutboxEvent is a transactional outbox record on the recoverable change
// stream. Cursor is a global position; PayloadVersion identifies the minimal
// payload schema.
type OutboxEvent struct {
	EventID        string         `json:"eventId"`
	EventType      string         `json:"eventType"`
	AggregateType  string         `json:"aggregateType"`
	AggregateID    string         `json:"aggregateId"`
	Sequence       int64          `json:"sequence"`
	Cursor         int64          `json:"cursor"`
	PayloadVersion int            `json:"payloadVersion"`
	Payload        map[string]any `json:"payload,omitempty"`
	OccurredAt     string         `json:"occurredAt"`
	Attempts       int            `json:"attempts"`
}

// EventPage is a page of the change stream.
type EventPage struct {
	Items      []OutboxEvent `json:"items"`
	NextCursor int64         `json:"nextCursor"`
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
		AgentRef    string   `json:"agentRef"`
		Name        string   `json:"name"`
		Purpose     string   `json:"purpose"`
		SponsorRef  string   `json:"sponsorRef"`
		OwnerRef    string   `json:"ownerRef"`
		RiskClass   string   `json:"riskClass"`
		AgentClass  string   `json:"agentClass,omitempty"`
		CarrierRefs []string `json:"carrierRefs,omitempty"`
	}{
		Mutation:    mut,
		AgentRef:    agent.AgentRef,
		Name:        agent.Name,
		Purpose:     agent.Purpose,
		SponsorRef:  agent.SponsorRef,
		OwnerRef:    agent.OwnerRef,
		RiskClass:   agent.RiskClass,
		AgentClass:  agent.AgentClass,
		CarrierRefs: agent.CarrierRefs,
	}
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

// GetRegistryContext returns a consistent point-in-time view for one Agent
// identity. Consumers making a lifecycle decision must use this rather than
// joining separate list responses.
func (c *Client) GetRegistryContext(ctx context.Context, namespace, agentID, instanceID, workloadRegistrationID string) (*RegistryContext, error) {
	query := url.Values{"namespace": {namespace}, "agentId": {agentID}}
	if instanceID != "" {
		query.Set("instanceId", instanceID)
	}
	if workloadRegistrationID != "" {
		query.Set("workloadRegistrationId", workloadRegistrationID)
	}
	var out RegistryContext
	if err := c.do(ctx, http.MethodGet, "/v1/registry-context?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
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
		Namespace           string             `json:"namespace"`
		Platform            string             `json:"platform"`
		Selector            map[string]string  `json:"selector"`
		TrustDomain         string             `json:"trustDomain"`
		AllowedProofMethods []string           `json:"allowedProofMethods"`
		ProofRequirements   *ProofRequirements `json:"proofRequirements,omitempty"`
	}{mut, workload.Namespace, workload.Platform, workload.Selector, workload.TrustDomain, workload.AllowedProofMethods, workload.ProofRequirements}
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
//
// IdempotencyKey is optional. Set it to a stable value for a logical enrollment
// so a retry replays the original instance instead of creating a second one; a
// retry with the same key and a changed payload is rejected with a conflict.
type InstanceCommit struct {
	Namespace              string    `json:"namespace"`
	WorkloadRegistrationID string    `json:"workloadRegistrationId"`
	WorkloadID             string    `json:"workloadId"`
	ArtifactDigest         string    `json:"artifactDigest"`
	AttestationRef         string    `json:"attestationRef"`
	LeaseExpiresAt         time.Time `json:"-"`
	IdempotencyKey         string    `json:"-"`
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
	if err := c.doKeyed(ctx, http.MethodPost, "/v1/agent-identities/"+agentID+"/instances", body, &out, commit.IdempotencyKey); err != nil {
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

// GetDiscovery resolves the registry discovery document for a namespace. The
// document is signed when the deployment configures a signing key.
func (c *Client) GetDiscovery(ctx context.Context, namespace string) (*DiscoveryDocument, error) {
	var out DiscoveryDocument
	if err := c.do(ctx, http.MethodGet, "/.well-known/agent-iam?namespace="+urlQueryEscape(namespace), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDiscoveryJWKS returns the registry discovery signing keys.
func (c *Client) GetDiscoveryJWKS(ctx context.Context) (*JWKS, error) {
	var out JWKS
	if err := c.do(ctx, http.MethodGet, "/.well-known/agent-iam/jwks.json", nil, &out); err != nil {
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

// ListEvents returns the change stream from the beginning.
func (c *Client) ListEvents(ctx context.Context) ([]OutboxEvent, error) {
	page, err := c.ReplayEvents(ctx, 0, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// ReplayEvents returns the change stream after a cursor. limit <= 0 means no bound.
func (c *Client) ReplayEvents(ctx context.Context, after int64, limit int) (*EventPage, error) {
	query := url.Values{"after": {strconv.FormatInt(after, 10)}}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	var out EventPage
	if err := c.do(ctx, http.MethodGet, "/v1/events?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LeaseEvents claims a batch of pending events for at-least-once delivery.
func (c *Client) LeaseEvents(ctx context.Context, owner string, limit, leaseSeconds int) (*EventPage, error) {
	body := map[string]any{"owner": owner, "limit": limit, "leaseSeconds": leaseSeconds}
	var out EventPage
	if err := c.do(ctx, http.MethodPost, "/v1/events/lease", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AckEvents marks leased events processed.
func (c *Client) AckEvents(ctx context.Context, owner string, cursors []int64) error {
	return c.do(ctx, http.MethodPost, "/v1/events/ack", map[string]any{"owner": owner, "cursors": cursors}, nil)
}

// NackEvents releases a lease for retry or dead-letters an exhausted event.
func (c *Client) NackEvents(ctx context.Context, owner string, cursors []int64, reason string, maxAttempts int) error {
	return c.do(ctx, http.MethodPost, "/v1/events/nack", map[string]any{
		"owner": owner, "cursors": cursors, "reason": reason, "maxAttempts": maxAttempts,
	}, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	return c.doKeyed(ctx, method, path, body, out, "")
}

// doKeyed is do with an explicit Idempotency-Key. An empty key generates one.
func (c *Client) doKeyed(ctx context.Context, method, path string, body any, out any, idempotencyKey string) error {
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
		key := idempotencyKey
		if key == "" {
			key = newIdempotencyKey()
		}
		request.Header.Set("Idempotency-Key", key)
		request.Header.Set("X-Actor", c.actor)
	}
	if c.tokenProvider != nil {
		token, err := c.tokenProvider()
		if err != nil {
			return fmt.Errorf("nomivela: resolve token: %w", err)
		}
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
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
