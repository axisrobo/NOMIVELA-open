# Status

Version `0.1.0` publishes the Core health and readiness contract plus the Agent Registry v0.1 registration and discovery contracts:

- `contracts/core-api.openapi.yaml`
- `contracts/agent-registry-v0.1.openapi.yaml` covering namespaces, agents, identities, bindings, workloads, instances, lifecycle, evidence, events, and discovery
- `contracts/schemas/agent-registry-v0.1.schema.json`
- `conformance/` manifest-driven fixtures plus a `go test ./conformance/...` suite that validates valid and invalid cases for Namespace, AgentRecord, AgentIdentity, and DiscoveryDocument
- `sdk/go` standard-library Go client for namespaces, agents, identities, bindings, workloads, instances, containment, evidence, and events, with httptest coverage
- `cli/nomivela` command-line client over the SDK, and `examples/go/quickstart` for the full registration and enrollment path
