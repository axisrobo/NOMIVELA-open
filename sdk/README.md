# NOMIVELA SDKs

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
