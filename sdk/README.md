# NOMIVELA SDKs

## Authentication

When the deployment enables API tokens or scoped service principals, send
`Authorization: Bearer <token>`. A scoped principal is restricted to the
namespaces it is configured for, and needs `registry.read` to read and
`instance.commit` to commit an instance.

```go
client := nomivela.New("https://registry.example.com", nomivela.WithToken(os.Getenv("NOMIVELA_TOKEN")))
```

```python
client = Client("https://registry.example.com", token=os.environ["NOMIVELA_TOKEN"])
```

```java
Client client = Client.withToken("https://registry.example.com", "eidovela", System.getenv("NOMIVELA_TOKEN"));
```

Use a token provider when the token rotates: `nomivela.WithTokenProvider(fn)`,
`Client(..., token_provider=fn)`, or `new Client(baseUrl, actor, supplier)`.

## Instance Commit Idempotency

Set a stable idempotency key for a logical enrollment so a retry replays the
original instance instead of creating a second one; the same key with a changed
payload is rejected with a conflict.

```go
instance, err := client.CommitInstance(ctx, agentID, nomivela.InstanceCommit{
    Namespace: "https://auth.example.com", WorkloadRegistrationID: regID,
    WorkloadID: "orders-1", ArtifactDigest: "sha256:abc", AttestationRef: "attestation:1",
    LeaseExpiresAt: time.Now().Add(time.Hour), IdempotencyKey: "enrollment-42",
}, nomivela.Mutation{})
```

```python
client.commit_instance(agent_id, commit, idempotency_key="enrollment-42")
```

```java
client.commitInstance(agentId, commit, mutation, "enrollment-42");
```

## Go

`sdk/go` is a standard-library-only client for the Agent Registry API. It sends
`Idempotency-Key` and `X-Actor` on every mutation, as the contract requires.

```go
client := nomivela.New("http://localhost:8080", nomivela.WithActor("user:platform"))

namespace, err := client.CreateNamespace(ctx, "https://auth.example.com", "root:org-a", nomivela.Mutation{Reason: "bootstrap"})
agent, err := client.CreateAgent(ctx, nomivela.Agent{
    AgentRef: "agent_order_processor", Name: "Order Processor", Purpose: "Process orders",
    SponsorRef: "org:platform", OwnerRef: "user:owner", RiskClass: "medium",
}, nomivela.Mutation{})

// One consistent read for an issuance decision.
context, err := client.GetRegistryContext(ctx, "https://auth.example.com", agentID, instanceID, "")
```

Run the tests:

```powershell
go test ./sdk/...
```

The client maps non-2xx responses to `*nomivela.APIError` carrying the stable
`Code` and `CorrelationID` from the contract.

## Python

`sdk/python` is a standard-library-only client (`urllib`) with the same behavior.

```python
from nomivela import Agent, Client, Mutation

client = Client("http://localhost:8080", actor="user:platform")
namespace = client.create_namespace("https://auth.example.com", "root:org-a", Mutation(reason="bootstrap"))
```

Run the tests:

```powershell
cd sdk/python; python -m unittest discover -s tests -t .
```

Non-2xx responses raise `nomivela.APIError` with `status`, `code`, `message`, and
`correlation_id`.

## Java

`sdk/java` is a zero-runtime-dependency client (JDK 17+) built with Maven.

```java
Client client = new Client("http://localhost:8080", "user:platform");
Namespace namespace = client.createNamespace("https://auth.example.com", "root:org-a",
        new Mutation("bootstrap", null));
```

Run the tests:

```powershell
cd sdk/java; mvn -o test
```

Non-2xx responses raise `ApiException` with `status()`, `code()`, and
`correlationId()`.
