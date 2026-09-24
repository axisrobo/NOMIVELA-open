# Development Harness

Superpowers is the required development harness for NOMIVELA Open.

## Versioning Rules

- Format: `major.minor.patch` (for example `0.1.0`, `1.4.2`). No pre-release or build suffixes.
- Open (`NOMIVELA-open`) and Core (`NOMIVELA`) share the same version number and the same tag `v<major>.<minor>.<patch>`, each tagged at its own repository commit.
- EE (`NOMIVELA-ee`) has an independent version and tag; it is not synchronized with this repository.
- Authoritative rule: the Core skill `.superpowers/skills/nomivela-release-versioning/SKILL.md`.
- Push tags only with explicit user confirmation.
