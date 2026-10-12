package nomivela

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func signedDiscovery(t *testing.T) (DiscoveryDocument, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	doc := DiscoveryDocument{
		DiscoveryVersion:       "1",
		Namespace:              "https://auth.example.com",
		RegistryEndpoint:       "https://registry.example.com",
		Issuer:                 "https://idp.example.com?a=1&b=2",
		JWKSUri:                "https://idp.example.com/jwks.json",
		SupportedProofProfiles: []string{"private_key_jwt"},
		IssuedAt:               "2026-01-02T03:04:05.123456789Z",
		ExpiresAt:              "2026-01-02T04:04:05Z",
		SigningKID:             "kid-1",
		Alg:                    "EdDSA",
	}
	doc.Signature = ""
	canonical, err := CanonicalDiscovery(doc)
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	doc.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv, canonical))
	return doc, pub
}

func TestVerifyDiscoveryAcceptsSignedDocument(t *testing.T) {
	doc, pub := signedDiscovery(t)
	if err := VerifyDiscovery(doc, pub); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerifyDiscoveryRejectsTamperAndWrongKey(t *testing.T) {
	doc, pub := signedDiscovery(t)

	tampered := doc
	tampered.RegistryEndpoint = "https://evil.example.com"
	if err := VerifyDiscovery(tampered, pub); err == nil {
		t.Fatal("tampered document verified")
	}

	otherPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	if err := VerifyDiscovery(doc, otherPub); err == nil {
		t.Fatal("document verified under the wrong key")
	}

	unsigned := doc
	unsigned.Signature = ""
	if err := VerifyDiscovery(unsigned, pub); err == nil {
		t.Fatal("unsigned document verified")
	}
}

func TestCanonicalDiscoveryIsReproducibleAndNormalizesTime(t *testing.T) {
	doc, _ := signedDiscovery(t)
	first, err := CanonicalDiscovery(doc)
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	second, _ := CanonicalDiscovery(doc)
	if !bytes.Equal(first, second) {
		t.Fatal("canonical bytes are not deterministic")
	}
	if strings.Contains(string(first), `\u0026`) || !strings.Contains(string(first), "&") {
		t.Fatalf("canonical must not HTML-escape: %s", first)
	}
	if !strings.Contains(string(first), `"issuedAt":"2026-01-02T03:04:05Z"`) {
		t.Fatalf("canonical must normalize timestamps to seconds: %s", first)
	}
}

func TestJWKPublicKeyAndVerifyWithJWKS(t *testing.T) {
	doc, pub := signedDiscovery(t)
	jwks := JWKS{Keys: []JWK{{
		Kty: "OKP", Crv: "Ed25519", Kid: "kid-1", Alg: "EdDSA", Use: "sig",
		X: base64.RawURLEncoding.EncodeToString(pub),
	}}}
	if err := VerifyDiscoveryWithJWKS(doc, jwks); err != nil {
		t.Fatalf("verify with jwks: %v", err)
	}

	decoded, err := JWKPublicKey(jwks.Keys[0])
	if err != nil {
		t.Fatalf("jwk public key: %v", err)
	}
	if !decoded.Equal(pub) {
		t.Fatal("decoded JWK key mismatch")
	}

	unknownKid := doc
	unknownKid.SigningKID = "kid-2"
	if err := VerifyDiscoveryWithJWKS(unknownKid, jwks); err == nil {
		t.Fatal("unknown kid must not verify")
	}
}
