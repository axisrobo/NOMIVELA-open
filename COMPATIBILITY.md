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

### Additive change: Agent Identity type + 1:N (2026-10-10)

The current `agent-registry-v2.0` line gains two optional Agent Identity fields —
`identityType` (`interactive`/`autonomous`/`digitalEmployee`) and `subjectRef` —
plus optional list filters, and relaxes uniqueness to one live identity per
`(agent, identityType, subjectRef)`. This is additive under the change policy
below: omitting the fields keeps the legacy 1:1 behavior, and a client that
ignores the new fields keeps working. The contract version and file names are
unchanged. New fixtures (`valid-agent-identity-typed`,
`invalid-agent-identity-type`, `invalid-agent-identity-subject-missing`,
`invalid-agent-identity-subject-forbidden`) exercise the pairing rule.

Contract `v1.0` was the graduation of the `v0.1` contract set at release `2.0.0`:
it froze the endpoint surface and field shapes that shipped across the
`1.0.0`-`1.8.0` releases, reached additively, so a `v0.1` client that ignores the
newer fields kept working. Contract `v2.0` (`2.0.1`) then applied the nine-class
camelCase Agent class vocabulary and is the current stable line; the `v1.0` line
is frozen.

### Additive: discovery verification in the SDKs

The Go, Python, and Java SDKs add discovery verification (`CanonicalDiscovery`,
`VerifyDiscovery`, `JWKPublicKey`, `VerifyDiscoveryWithJWKS`; Python exports
`canonical_discovery`/`verify_discovery`/`jwk_public_key`/
`verify_discovery_with_jwks`; Java exposes `DiscoveryVerification`) so a consumer
verifies a signed document against an independently trusted key without
reimplementing the canonical form or coupling to Core internals. Additive and
SDK-only; the contract's endpoint surface and field shapes are unchanged. The
canonical form is documented on `DiscoveryDocument.signature`: sorted keys, HTML
escaping disabled, RFC 3339 UTC timestamps at second precision. All three SDKs
are checked against one shared cross-language vector.

## Repository Alignment

- `NOMIVELA` (Core, AGPL-3.0) and `NOMIVELA-open` always share the same release
  version and tag `v<major>.<minor>.<patch>`.
- `NOMIVELA-ee` uses an independent version and tag.
- See the Core skill `.superpowers/skills/nomivela-release-versioning/SKILL.md`.

## Release Compatibility Matrix

| Release | Core / Open | Contracts | SDKs | Notes |
| --- | --- | --- | --- | --- |
| `v2.1.0` | `2.1.0` | `agent-registry-v2.0`, `eidovela-v1-compat` | Go, Python, Java | Agent Identity type (`identityType`/`subjectRef`) + scoped 1:N uniqueness; additive, no contract-version change |
| `v2.0.1` | `2.0.1` | `agent-registry-v2.0`, `eidovela-v1-compat` | Go, Python, Java | Contract `v2.0`: nine-class camelCase Agent class vocabulary |
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
