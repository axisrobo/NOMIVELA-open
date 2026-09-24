# NOMIVELA Python SDK

A standard-library-only client for the Agent Registry API. It sends
`Idempotency-Key` and `X-Actor` on every mutation and raises `APIError` with the
stable contract code on failure.

```python
from nomivela import Agent, Client, Mutation

client = Client("http://localhost:8080", actor="user:platform")
mutation = Mutation(reason="bootstrap")

namespace = client.create_namespace("https://auth.example.com", "root:org-a", mutation)
agent = client.create_agent(
    Agent(
        agent_ref="agent_edge_twin",
        name="Edge Twin",
        purpose="Mirror the edge gateway",
        sponsor_ref="org:platform",
        owner_ref="user:owner",
        risk_class="high",
        agent_class="asset_twin",
        carrier_refs=["device:gateway-7"],
    ),
    mutation,
)
```

## Tests

```powershell
python -m unittest discover -s tests -t .
```
