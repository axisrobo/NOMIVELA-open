# NOMIVELA Open Roadmap

## Roadmap Principle

Open publishes contracts before Core implements them and distributes artifacts only after the matching Core release passes conformance. No contract is released without a schema, a positive fixture, and a negative fixture.

Open carries the public surface: OpenAPI, JSON Schema, conformance fixtures, SDKs, CLI, examples, compatibility adapters, and redistributable Core binaries. It never carries Core source, Enterprise controls, private deployment configuration, or internal authority mappings.

## O0: Contract Foundation (0.1.0) - complete

**Deliverables**

- Core health and readiness contract.
- Agent Registry v0.1 OpenAPI and JSON Schema for Namespace, Agent Record, Agent Identity, Authority Binding, Workload Registration, Agent Instance, Discovery Document, lifecycle event, and outbox event.
- Conformance fixtures with positive and negative Agent Record cases.
- Public Go module and repository structure.

**Exit gate**

- Schema and fixtures validate; invalid `agentRef` is rejected by contract.

## O1: Developer Preview Contracts (0.5.0) - in progress

**Deliverables**

- Complete versioned contracts for namespace, Agent, identity, immutable binding, workload, instance, discovery, evidence, and events. Published in `contracts/agent-registry-v0.1.openapi.yaml`. Done.
- Schema-driven conformance fixtures and a `go test ./conformance/...` suite covering valid and invalid cases for Namespace, AgentRecord, AgentIdentity, and DiscoveryDocument. Done.
- Go SDK for the registration, lifecycle, containment, and evidence API in `sdk/go`. Done.
- Public CLI (`cli/nomivela`) for namespace, Agent, identity, workload, instance, evidence, and event workflows. Done.
- Runnable integration example (`examples/go/quickstart`). Done.
- EIDOVELA v1 compatibility fixtures and schema (`contracts/schemas/eidovela-v1-compat.schema.json`). Done.

**Exit gate**

- SDK, CLI, and examples complete the same registration and discovery path exercised by Core conformance.

## O2: Multi-Language SDKs And Full Conformance (1.0.0)

**Deliverables**

- Go SDK and CLI covering the registration, lifecycle, containment, evidence, and discovery API. Done.
- Conformance suite covering schema-valid and schema-invalid cases, plus Core-side behavioral conformance for uniqueness, immutability, epochs, lifecycle, discovery, and events. Done.
- Python SDK (`sdk/python`, standard library only). Done (post-`1.0.0`).
- Java SDK (`sdk/java`, zero runtime dependencies, JDK 17+). Done (post-`1.0.0`).
- Binary checksums and SBOM per release: Core produces them with `scripts/release.ps1` (`SHA256SUMS`, manifest, CycloneDX `sbom.json`) and they are published as GitHub Release assets on this repository. Done (post-`1.0.0`).
- Compatibility documentation per release. Pending.

**Exit gate**

- Every supported SDK passes the shared conformance suite against a released Core binary. At `v1.1.0` the supported SDKs are Go, Python, and Java, each covered by its own test suite.

## O3: Federation And Edge Contracts (1.5.0, 2.0.0)

**Deliverables**

- Federation and SCIM contract profiles.
- Signed discovery and trust-bundle contracts.
- Edge projection contract for read-only regional consumers.

**Exit gate**

- Degraded and revoked trust fixtures fail closed in every SDK.
