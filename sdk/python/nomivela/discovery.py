"""Discovery document verification for the NOMIVELA Python SDK.

A consumer fetches the discovery document and verifies its Ed25519 signature
against a key it trusts independently (its own pinned key, or a key from a JWKS
it already trusts). This module reproduces the registry's canonical form exactly,
so verification needs no Core internals and no third-party dependency (the
Ed25519 check is implemented on the standard library).

Verifying a signature does not establish freshness; the caller must also enforce
``issuedAt``/``expiresAt`` and its namespace/endpoint policy.
"""

from __future__ import annotations

import base64
import hashlib
import json
from datetime import datetime, timezone
from typing import Any, Optional


class DiscoveryVerificationError(ValueError):
    """A discovery document failed verification."""


def _b64url_decode(value: str) -> bytes:
    padded = value + "=" * (-len(value) % 4)
    return base64.urlsafe_b64decode(padded.encode("ascii"))


def _canonical_time(value: Optional[str]) -> str:
    if not value:
        return ""
    parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
    # The registry signs second-precision UTC; normalize so a verifier that
    # re-serializes the emitted document reproduces the signed bytes.
    return parsed.astimezone(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def canonical_discovery(document: Any) -> bytes:
    """Return the deterministic bytes the registry signs (sorted keys, no HTML
    escaping, second-precision RFC 3339 UTC timestamps)."""
    payload = {
        "discoveryVersion": document.discovery_version or "",
        "namespace": document.namespace or "",
        "registryEndpoint": document.registry_endpoint or "",
        "issuer": document.issuer or "",
        "jwksUri": document.jwks_uri or "",
        "supportedProofProfiles": document.supported_proof_profiles,
        "supportedArtifactTypes": document.supported_artifact_types,
        "keyRotation": document.key_rotation or "",
        "issuedAt": _canonical_time(document.issued_at),
        "expiresAt": _canonical_time(document.expires_at),
        "signingKid": document.signing_kid or "",
        "alg": document.alg or "",
    }
    return json.dumps(payload, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode("utf-8")


def jwk_public_key(jwk: dict[str, Any]) -> bytes:
    """Decode an Ed25519 verification JWK into a raw 32-byte public key."""
    if jwk.get("kty") != "OKP" or jwk.get("crv") != "Ed25519":
        raise DiscoveryVerificationError(f"unsupported JWK kty={jwk.get('kty')} crv={jwk.get('crv')}")
    raw = _b64url_decode(str(jwk.get("x", "")))
    if len(raw) != 32:
        raise DiscoveryVerificationError(f"JWK x must be 32 bytes, got {len(raw)}")
    return raw


def verify_discovery(document: Any, public_key: bytes) -> None:
    """Verify ``document.signature`` over the canonical document. Raises
    DiscoveryVerificationError; does not enforce the validity window or trust."""
    signature = document.signature
    if not signature:
        raise DiscoveryVerificationError("discovery document is not signed")
    if len(public_key) != 32:
        raise DiscoveryVerificationError(f"verification key must be 32 bytes, got {len(public_key)}")
    raw_signature = _b64url_decode(signature)
    if not _ed25519_verify(public_key, canonical_discovery(document), raw_signature):
        raise DiscoveryVerificationError("discovery signature does not verify")


def verify_discovery_with_jwks(document: Any, jwks: Any) -> None:
    """Verify using the JWKS key named by ``signingKid``. Trust the JWKS
    independently; fetching it from the same origin as the document does not
    establish trust."""
    for jwk in jwks.keys or []:
        if jwk.get("kid") != document.signing_kid:
            continue
        verify_discovery(document, jwk_public_key(jwk))
        return
    raise DiscoveryVerificationError(f"no JWKS key for signing kid {document.signing_kid!r}")


# --- Ed25519 verification (RFC 8032), standard library only -------------------

_Q = 2**255 - 19
_L = 2**252 + 27742317777372353535851937790883648493


def _expmod(base: int, exponent: int, modulus: int) -> int:
    return pow(base, exponent, modulus)


def _inv(x: int) -> int:
    return _expmod(x, _Q - 2, _Q)


_D = -121665 * _inv(121666) % _Q
_I = _expmod(2, (_Q - 1) // 4, _Q)


def _xrecover(y: int) -> int:
    xx = (y * y - 1) * _inv(_D * y * y + 1) % _Q
    x = _expmod(xx, (_Q + 3) // 8, _Q)
    if (x * x - xx) % _Q != 0:
        x = (x * _I) % _Q
    if x % 2 != 0:
        x = _Q - x
    return x


_BY = 4 * _inv(5) % _Q
_B = [_xrecover(_BY) % _Q, _BY % _Q]


def _edwards(p: list[int], q: list[int]) -> list[int]:
    x1, y1 = p
    x2, y2 = q
    x3 = (x1 * y2 + x2 * y1) * _inv(1 + _D * x1 * x2 * y1 * y2) % _Q
    y3 = (y1 * y2 + x1 * x2) * _inv(1 - _D * x1 * x2 * y1 * y2) % _Q
    return [x3, y3]


def _scalarmult(p: list[int], e: int) -> list[int]:
    if e == 0:
        return [0, 1]
    q = _scalarmult(p, e // 2)
    q = _edwards(q, q)
    if e & 1:
        q = _edwards(q, p)
    return q


def _decodepoint(s: bytes) -> list[int]:
    y = int.from_bytes(s, "little") & ((1 << 255) - 1)
    x = _xrecover(y)
    if (x & 1) != (s[31] >> 7):
        x = _Q - x
    return [x, y]


def _ed25519_verify(public_key: bytes, message: bytes, signature: bytes) -> bool:
    if len(signature) != 64 or len(public_key) != 32:
        return False
    r = _decodepoint(signature[:32])
    s = int.from_bytes(signature[32:], "little")
    if s >= _L:
        return False
    a = _decodepoint(public_key)
    h = int.from_bytes(hashlib.sha512(signature[:32] + public_key + message).digest(), "little") % _L
    return _scalarmult(_B, s) == _edwards(r, _scalarmult(a, h))
