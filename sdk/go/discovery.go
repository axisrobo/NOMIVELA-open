package nomivela

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Discovery verification lets a consumer trust the registry's published
// discovery document using only the public contract — no Core internals. A
// caller fetches the document (GetDiscovery) and verifies Signature against a
// public key it trusts independently (its own pinned key, or a key from a JWKS
// it already trusts). Verifying a signature does not establish freshness; the
// caller must also enforce IssuedAt/ExpiresAt and its namespace/endpoint policy.

// CanonicalDiscovery returns the deterministic bytes the registry signs, so an
// independent verifier can reproduce them. Keys are sorted, HTML escaping is
// disabled, and timestamps are RFC 3339 in UTC at second precision.
func CanonicalDiscovery(doc DiscoveryDocument) ([]byte, error) {
	issuedAt, err := canonicalTime(doc.IssuedAt)
	if err != nil {
		return nil, fmt.Errorf("nomivela: issuedAt: %w", err)
	}
	expiresAt, err := canonicalTime(doc.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("nomivela: expiresAt: %w", err)
	}
	payload := map[string]any{
		"discoveryVersion":       doc.DiscoveryVersion,
		"namespace":              doc.Namespace,
		"registryEndpoint":       doc.RegistryEndpoint,
		"issuer":                 doc.Issuer,
		"jwksUri":                doc.JWKSUri,
		"supportedProofProfiles": doc.SupportedProofProfiles,
		"supportedArtifactTypes": doc.SupportedArtifactTypes,
		"keyRotation":            doc.KeyRotation,
		"issuedAt":               issuedAt,
		"expiresAt":              expiresAt,
		"signingKid":             doc.SigningKID,
		"alg":                    doc.Alg,
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func canonicalTime(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", err
	}
	// The registry signs second-precision UTC; normalize so a verifier that
	// re-serializes the emitted document reproduces the signed bytes.
	return parsed.UTC().Format(time.RFC3339), nil
}

// VerifyDiscovery checks doc.Signature over CanonicalDiscovery(doc) against key.
// It does not enforce the validity window or trust the key; the caller must
// supply a key it trusts and apply its own clock and namespace policy.
func VerifyDiscovery(doc DiscoveryDocument, key ed25519.PublicKey) error {
	if doc.Signature == "" {
		return errors.New("nomivela: discovery document is not signed")
	}
	if len(key) != ed25519.PublicKeySize {
		return fmt.Errorf("nomivela: verification key must be %d bytes, got %d", ed25519.PublicKeySize, len(key))
	}
	signature, err := base64.RawURLEncoding.DecodeString(doc.Signature)
	if err != nil {
		return fmt.Errorf("nomivela: decode signature: %w", err)
	}
	input, err := CanonicalDiscovery(doc)
	if err != nil {
		return err
	}
	if !ed25519.Verify(key, input, signature) {
		return errors.New("nomivela: discovery signature does not verify")
	}
	return nil
}

// JWKPublicKey decodes an Ed25519 verification JWK into a public key.
func JWKPublicKey(jwk JWK) (ed25519.PublicKey, error) {
	if jwk.Kty != "OKP" || jwk.Crv != "Ed25519" {
		return nil, fmt.Errorf("nomivela: unsupported JWK kty=%q crv=%q", jwk.Kty, jwk.Crv)
	}
	raw, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("nomivela: decode JWK x: %w", err)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("nomivela: JWK x must be %d bytes, got %d", ed25519.PublicKeySize, len(raw))
	}
	return ed25519.PublicKey(raw), nil
}

// VerifyDiscoveryWithJWKS verifies a document using the JWKS key named by its
// SigningKID. Trust the JWKS independently; fetching it from the same origin as
// the document does not establish trust.
func VerifyDiscoveryWithJWKS(doc DiscoveryDocument, jwks JWKS) error {
	for _, jwk := range jwks.Keys {
		if jwk.Kid != doc.SigningKID {
			continue
		}
		key, err := JWKPublicKey(jwk)
		if err != nil {
			return err
		}
		return VerifyDiscovery(doc, key)
	}
	return fmt.Errorf("nomivela: no JWKS key for signing kid %q", doc.SigningKID)
}
