import base64
import unittest

from nomivela import DiscoveryVerificationError
from nomivela import canonical_discovery, jwk_public_key, verify_discovery, verify_discovery_with_jwks
from nomivela.types import DiscoveryDocument, JWKS

# Cross-language vector: signed by the Go SDK's canonical form with seed 0x01*32.
CANONICAL = (
    '{"alg":"EdDSA","discoveryVersion":"1","expiresAt":"2026-01-02T04:04:05Z",'
    '"issuedAt":"2026-01-02T03:04:05Z","issuer":"https://idp.example.com",'
    '"jwksUri":"https://idp.example.com/jwks.json","keyRotation":"30d",'
    '"namespace":"https://auth.example.com","registryEndpoint":"https://registry.example.com",'
    '"signingKid":"kid-vector","supportedArtifactTypes":["x509"],'
    '"supportedProofProfiles":["private_key_jwt"]}'
)
SIGNATURE = "_LLAb1V-a3X-Z9Ga-3eoZ9bjlNP8bCUVbRiihI6UhlLNCluzl2A7kyeikvNL_iwO7XM-9L654OB5kbv7ordmDQ"
PUBLIC_KEY = base64.urlsafe_b64decode("iojj3XQJ8ZX9UtstPLpdcspnCb8dlBIb83SIAbQPb1w" + "==")


def document() -> DiscoveryDocument:
    return DiscoveryDocument(
        namespace="https://auth.example.com",
        registry_endpoint="https://registry.example.com",
        issuer="https://idp.example.com",
        jwks_uri="https://idp.example.com/jwks.json",
        supported_proof_profiles=["private_key_jwt"],
        supported_artifact_types=["x509"],
        key_rotation="30d",
        discovery_version="1",
        issued_at="2026-01-02T03:04:05Z",
        expires_at="2026-01-02T04:04:05Z",
        signing_kid="kid-vector",
        alg="EdDSA",
        signature=SIGNATURE,
    )


class DiscoveryVerificationTest(unittest.TestCase):
    def test_canonical_matches_cross_language_vector(self):
        self.assertEqual(CANONICAL, canonical_discovery(document()).decode("utf-8"))

    def test_canonical_normalizes_fractional_seconds(self):
        doc = document()
        doc.issued_at = "2026-01-02T03:04:05.123456Z"
        self.assertIn('"issuedAt":"2026-01-02T03:04:05Z"', canonical_discovery(doc).decode("utf-8"))

    def test_verify_accepts_signed_document(self):
        verify_discovery(document(), PUBLIC_KEY)

    def test_verify_rejects_tamper_wrong_key_and_unsigned(self):
        tampered = document()
        tampered.registry_endpoint = "https://evil.example.com"
        with self.assertRaises(DiscoveryVerificationError):
            verify_discovery(tampered, PUBLIC_KEY)

        with self.assertRaises(DiscoveryVerificationError):
            verify_discovery(document(), bytes(32))

        unsigned = document()
        unsigned.signature = None
        with self.assertRaises(DiscoveryVerificationError):
            verify_discovery(unsigned, PUBLIC_KEY)

    def test_verify_with_jwks(self):
        jwks = JWKS(keys=[{
            "kty": "OKP", "crv": "Ed25519", "kid": "kid-vector", "alg": "EdDSA", "use": "sig",
            "x": "iojj3XQJ8ZX9UtstPLpdcspnCb8dlBIb83SIAbQPb1w",
        }])
        verify_discovery_with_jwks(document(), jwks)
        self.assertEqual(PUBLIC_KEY, jwk_public_key(jwks.keys[0]))

        unknown = document()
        unknown.signing_kid = "kid-2"
        with self.assertRaises(DiscoveryVerificationError):
            verify_discovery_with_jwks(unknown, jwks)


if __name__ == "__main__":
    unittest.main()
