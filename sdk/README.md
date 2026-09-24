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
