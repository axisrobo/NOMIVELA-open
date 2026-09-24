# Status

Version `1.0.0` is the stable public NOMIVELA surface: contracts, schemas, fixtures, the Go SDK, CLI, and examples.

## Published

- `contracts/core-api.openapi.yaml`
- `contracts/agent-registry-v0.1.openapi.yaml`: namespaces, agents, identities, bindings, workloads, instances, lifecycle, containment, evidence, events, and discovery (17 paths, 20 schemas)
- `contracts/schemas/agent-registry-v0.1.schema.json`, including the Agent form (`agentClass`) and carrier references (`carrierRefs`) introduced by ADR 0005
- `contracts/schemas/eidovela-v1-compat.schema.json`
- `conformance/`: manifest-driven fixtures and a `go test ./conformance/...` suite validating valid and invalid cases across both schemas
- `sdk/go`: standard-library Go client for namespaces, agents, identities, bindings, workloads, instances, containment, evidence, and events
- `sdk/python`: standard-library Python client with the same surface
- `sdk/java`: zero-runtime-dependency Java client (JDK 17+, Maven) with the same surface
- `cli/nomivela`: command-line client over the SDK
- `examples/go/quickstart`: the full registration and enrollment path

Post-`1.0.0` on `main`: the Python (`sdk/python`) and Java (`sdk/java`) SDKs are added and tested. Core produces release binaries with `SHA256SUMS` and a CycloneDX SBOM via `scripts/release.ps1`; publishing those artifacts here is the remaining step. At the `v1.0.0` tag the Go SDK is the only supported SDK.

See `docs/roadmap.md` for the delivery plan.
