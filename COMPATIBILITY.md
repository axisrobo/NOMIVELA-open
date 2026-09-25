# Compatibility

## Two Version Numbers

Do not confuse them.

| Version | Example | Meaning |
| --- | --- | --- |
| Release version | `1.1.0` | The product release. Core and Open share it; EE is independent. |
| Contract version | `0.1` | The Agent Registry public contract set (`agent-registry-v0.1`, `eidovela-v1-compat`). Changes only when a documented compatibility decision allows it. |

A release may add optional fields to a contract without changing the contract
version. The `v0.1` contract file names and `info.version` stay fixed until an
approved contract-version bump, which is guarded by
`conformance.TestContractVersionIsStable`.

## Repository Alignment

- `NOMIVELA` (Core, AGPL-3.0) and `NOMIVELA-open` always share the same release
  version and tag `v<major>.<minor>.<patch>`.
- `NOMIVELA-ee` uses an independent version and tag.
- See the Core skill `.superpowers/skills/nomivela-release-versioning/SKILL.md`.

## Release Compatibility Matrix

| Release | Core / Open | Contracts | SDKs | Notes |
| --- | --- | --- | --- | --- |
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
