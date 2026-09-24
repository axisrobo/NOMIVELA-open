# Status

Version `0.5.0` is the Developer Preview of the public NOMIVELA surface: contracts, schemas, fixtures, SDK, CLI, and examples.

## Published

- `contracts/core-api.openapi.yaml`
- `contracts/agent-registry-v0.1.openapi.yaml`: namespaces, agents, identities, bindings, workloads, instances, lifecycle, containment, evidence, events, and discovery (17 paths, 20 schemas)
- `contracts/schemas/agent-registry-v0.1.schema.json`
- `contracts/schemas/eidovela-v1-compat.schema.json`
- `conformance/`: manifest-driven fixtures and a `go test ./conformance/...` suite validating valid and invalid cases across both schemas
- `sdk/go`: standard-library Go client for namespaces, agents, identities, bindings, workloads, instances, containment, evidence, and events
- `cli/nomivela`: command-line client over the SDK
- `examples/go/quickstart`: the full registration and enrollment path

See `docs/roadmap.md` for the delivery plan.
