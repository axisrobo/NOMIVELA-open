# Compatibility

## Two Version Numbers

Do not confuse them.

| Version | Example | Meaning |
| --- | --- | --- |
| Release version | `2.0.0` | The product release. Core and Open share it; EE is independent. |
| Contract version | `v2.0` | The Agent Registry public contract set (`agent-registry-v2.0`, `eidovela-v1-compat`). Changes only when a documented compatibility decision allows it. |

A release may add optional fields to a contract without changing the contract
version. The `v2.0` contract file names and `info.version` stay fixed until an
approved contract-version bump, which is guarded by
`conformance.TestContractVersionIsStable`.

`agent-registry-v2.0` is the current line. It applies the Agent IAM Series
contract conventions (RFC-0003): the Agent class values are `lowerCamelCase` and
the set is the unified nine-class vocabulary (`embedded`, `organizational`,
`user`, `assetTwin`, `personalTwin`, `twin`, `service`, `ephemeral`,
`simulation`). The previous `agent-registry-v1.0` line is frozen; it used
`asset_twin` and `personal_twin`.

Contract `v1.0` is the graduation of the `v0.1` contract set: it freezes the
endpoint surface and field shapes that shipped across the `1.0.0`-`1.8.0`
releases. It was reached additively, so a `v0.1` client that ignores the newer
fields keeps working; the file rename to `agent-registry-v1.0` is the intentional
breaking change that marks the stable line.

## Repository Alignment

- `NOMIVELA` (Core, AGPL-3.0) and `NOMIVELA-open` always share the same release
  version and tag `v<major>.<minor>.<patch>`.
- `NOMIVELA-ee` uses an independent version and tag.
- See the Core skill `.superpowers/skills/nomivela-release-versioning/SKILL.md`.

## Release Compatibility Matrix

| Release | Core / Open | Contracts | SDKs | Notes |
| --- | --- | --- | --- | --- |
| `v2.0.0` | `2.0.0` | `agent-registry-v1.0`, `eidovela-v1-compat` | Go, Python, Java | Contract `v1.0` graduation: stable endpoint and field surface |
| `v1.8.0` | `1.8.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Versioned workload proof profile (`proofRequirements`) |
| `v1.7.0` | `1.7.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Conditional reads (ETag/If-None-Match); paginated collections; expected-epoch/If-Match optimistic concurrency |
| `v1.6.0` | `1.6.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Recoverable event stream: cursor replay, lease/ack/nack; SDK event methods |
| `v1.5.0` | `1.5.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Signed discovery; registry discovery JWKS; SDK discovery reads |
| `v1.4.0` | `1.4.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Idempotent instance commit; SDK idempotency key |
| `v1.3.0` | `1.3.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Atomic Registry Context point read; scoped service principals; SDK bearer tokens |
| `v1.2.0` | `1.2.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Additive external evidence ingest (`/v1/evidence/external`) |
| `v1.1.0` | `1.1.0` | `agent-registry-v0.1`, `eidovela-v1-compat` | Go, Python, Java | Agent form and carrier model; Core binaries with checksums and SBOM |
| `v1.0.0` | `1.0.0` | `agent-registry-v0.1` | Go | Production baseline |
| `v0.5.0` | `0.5.0` | `agent-registry-v0.1` | Go | Developer Preview |
| `v0.1.0` | `0.1.0` | `agent-registry-v0.1` | none | Contract foundation |

## Change Policy

- **Additive within a contract version**: new optional fields, new endpoints, and
  new enums values are allowed and are exercised by conformance fixtures.
- **Breaking**: removing or renaming a field, tightening validation, or changing
  a state machine requires a contract-version bump and a product major release.
- **Deprecation**: a deprecated field or endpoint is supported for at least one
  minor release and documented in the release notes before removal.
- **SDK compatibility**: an SDK release targets one contract version. A newer
  SDK remains compatible with an older compatible Core release within the same
  contract version.

## Runtime Compatibility

| Component | Supported |
| --- | --- |
| Core runtime | Go 1.25 or later |
| Database | PostgreSQL 16 through 18 |
| SDK runtimes | Go 1.25, Python 3.9+, Java 17+ |
| API transport | HTTPS in production (HTTP allowed for local development) |
