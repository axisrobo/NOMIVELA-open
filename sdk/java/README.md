# NOMIVELA Java SDK

A zero-runtime-dependency Java client (JDK 17+) for the Agent Registry API. It
uses `java.net.http.HttpClient` and a small built-in JSON reader/writer.

```java
Client client = new Client("http://localhost:8080", "user:platform");

Namespace namespace = client.createNamespace("https://auth.example.com", "root:org-a",
        new Mutation("bootstrap", null));

Agent agent = client.createAgent(new Agent(
        "agent_edge_twin", "Edge Twin", "Mirror the edge gateway",
        "org:platform", "user:owner", "high",
        "asset_twin", List.of("device:gateway-7"), null, null),
        Mutation.none());
```

Every mutation sends `Idempotency-Key` and `X-Actor`. Non-2xx responses raise
`ApiException` with `status()`, `code()`, and `correlationId()`.

## Build and test

```powershell
cd sdk/java
mvn -o test   # offline
mvn test      # with network access
```
