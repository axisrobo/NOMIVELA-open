# Status

Version `1.2.0` is the current release of the public NOMIVELA surface: contracts, schemas, fixtures, the Go, Python, and Java SDKs, the CLI, examples, and redistributable Core binaries on the release page. `1.0.0` established the stable Go surface; `1.1.0` added the Python and Java SDKs and published Core release artifacts; `1.2.0` adds the external evidence ingest contract.

## Published

- `contracts/core-api.openapi.yaml`
- `contracts/agent-registry-v0.1.openapi.yaml`: namespaces, agents, identities, bindings, workloads, instances, lifecycle, containment, evidence, external evidence ingest, events, and discovery (18 paths, 23 schemas)
- `contracts/schemas/agent-registry-v0.1.schema.json`, including the Agent form (`agentClass`) and carrier references (`carrierRefs`) introduced by ADR 0005
- `contracts/schemas/eidovela-v1-compat.schema.json`
- `conformance/`: manifest-driven fixtures and a `go test ./conformance/...` suite validating valid and invalid cases across both schemas
- `sdk/go`: standard-library Go client for namespaces, agents, identities, bindings, workloads, instances, containment, evidence, and events
- `sdk/python`: standard-library Python client with the same surface
- `sdk/java`: zero-runtime-dependency Java client (JDK 17+, Maven) with the same surface
- `cli/nomivela`: command-line client over the SDK
- `examples/go/quickstart`: the full registration and enrollment path

At `v1.2.0` the supported SDKs are Go, Python, and Java. Core release binaries with `SHA256SUMS`, a release manifest, and a CycloneDX SBOM are published as GitHub Release assets on this repository by `scripts/release.ps1` in the core repository.

Compatibility between releases and contracts is documented in `COMPATIBILITY.md`.

See `docs/roadmap.md` for the delivery plan.
