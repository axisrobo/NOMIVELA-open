# Status

Version `2.0.0` is the current release of the public NOMIVELA surface: contracts, schemas, fixtures, the Go, Python, and Java SDKs, the CLI, examples, and redistributable Core binaries on the release page. `1.0.0` established the stable Go surface; `1.1.0` added the Python and Java SDKs and published Core release artifacts; `1.2.0` added the external evidence ingest contract; `1.3.0` added the atomic Registry Context contract and bearer-token support across all SDKs; `1.4.0` documented idempotent instance commit; `1.5.0` added signed discovery and the registry discovery JWKS; `1.6.0` added the recoverable event stream (cursor replay, lease/ack/nack); `1.7.0` added conditional reads (ETag), cursor-paginated collections, and expected-epoch optimistic concurrency; `1.8.0` added the versioned workload proof profile (`proofRequirements`); `2.0.0` graduates the Agent Registry contract from `v0.1` to `v1.0` (`agent-registry-v1.0`).

## Published

- `contracts/core-api.openapi.yaml`
- `contracts/agent-registry-v1.0.openapi.yaml`: namespaces, agents, identities, bindings, workloads, instances, lifecycle, containment, evidence, external evidence ingest, the atomic Registry Context point read, events, the recoverable event stream, signed discovery with the registry JWKS, the versioned workload proof profile, and schema definitions (23 paths, 34 schemas)
- `contracts/schemas/agent-registry-v1.0.schema.json`, including the Agent form (`agentClass`) and carrier references (`carrierRefs`) introduced by ADR 0005
- `contracts/schemas/eidovela-v1-compat.schema.json`
- `conformance/`: manifest-driven fixtures and a `go test ./conformance/...` suite validating valid and invalid cases across both schemas
- `sdk/go`: standard-library Go client for namespaces, agents, identities, bindings, workloads, instances, containment, evidence, and events
- `sdk/python`: standard-library Python client with the same surface
- `sdk/java`: zero-runtime-dependency Java client (JDK 17+, Maven) with the same surface
- `cli/nomivela`: command-line client over the SDK
- `examples/go/quickstart`: the full registration and enrollment path

At `v2.0.0` the supported SDKs are Go, Python, and Java, and all three cover the consumer surface: the atomic Registry Context read, signed discovery and the registry JWKS, and the recoverable event stream (cursor replay plus lease/ack/nack), in addition to registration and lifecycle. Core release binaries with `SHA256SUMS`, a release manifest, and a CycloneDX SBOM are published as GitHub Release assets on this repository by `scripts/release.ps1` in the core repository.

Compatibility between releases and contracts is documented in `COMPATIBILITY.md`.

See `docs/roadmap.md` for the delivery plan.
