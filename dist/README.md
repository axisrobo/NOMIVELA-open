# NOMIVELA Binaries

Prebuilt NOMIVELA Core binaries are published on the
[Releases](https://github.com/axisrobo/NOMIVELA-open/releases) page of this
repository. Each release tag (`v0.1.0`, `v0.5.0`, `v1.0.0`, `v1.1.0`, ...)
contains the Core commands for linux, darwin, and windows on amd64 and arm64.

The binaries are built from the AGPL-3.0
[NOMIVELA](https://github.com/axisrobo/NOMIVELA) core repository with
`scripts/release.ps1`, which injects the build version and produces
`SHA256SUMS`, a release manifest, and a CycloneDX SBOM. They remain subject to
AGPL-3.0.

## Contents of each release

- `nomivela-<os>-<arch>` - Core HTTP service
- `nomivela-migrate-<os>-<arch>` - migration runner (`-reset` is destructive and for development databases only)
- `nomivela-verify-<os>-<arch>` - registry invariant verifier
- `SHA256SUMS` - checksums for every binary
- `nomivela-v<version>-manifest.json` - artifact metadata
- `sbom.json` - CycloneDX 1.5 software bill of materials

Verify a download:

```powershell
$expected = (Select-String -Path SHA256SUMS -Pattern "nomivela-linux-amd64$").Line.Split(" ")[0]
(Get-FileHash -Algorithm SHA256 nomivela-linux-amd64).Hash.ToLowerInvariant() -eq $expected
```
