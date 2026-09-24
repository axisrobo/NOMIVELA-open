# Agent Registry Conformance

This package validates the published contract fixtures against the published JSON Schema.

- `fixtures/manifest.json` lists each fixture with its target `$defs` definition and whether it must be accepted or rejected.
- `fixtures/valid-*.json` must validate against the definition.
- `fixtures/invalid-*.json` must be rejected by the definition.

Run the suite:

```powershell
go test ./conformance/...
```

Add a fixture and a manifest entry with every published contract change. Behavioral rules that a JSON Schema cannot express (uniqueness, immutability, fail-closed discovery, containment cascade) are covered by the Core test suite.
