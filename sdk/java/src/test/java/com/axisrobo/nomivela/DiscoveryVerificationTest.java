package com.axisrobo.nomivela;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.Test;

class DiscoveryVerificationTest {

    // Cross-language vector: signed by the Go SDK's canonical form with seed 0x01*32.
    private static final String CANONICAL =
            "{\"alg\":\"EdDSA\",\"discoveryVersion\":\"1\",\"expiresAt\":\"2026-01-02T04:04:05Z\","
            + "\"issuedAt\":\"2026-01-02T03:04:05Z\",\"issuer\":\"https://idp.example.com\","
            + "\"jwksUri\":\"https://idp.example.com/jwks.json\",\"keyRotation\":\"30d\","
            + "\"namespace\":\"https://auth.example.com\",\"registryEndpoint\":\"https://registry.example.com\","
            + "\"signingKid\":\"kid-vector\",\"supportedArtifactTypes\":[\"x509\"],"
            + "\"supportedProofProfiles\":[\"private_key_jwt\"]}";
    private static final String SIGNATURE =
            "_LLAb1V-a3X-Z9Ga-3eoZ9bjlNP8bCUVbRiihI6UhlLNCluzl2A7kyeikvNL_iwO7XM-9L654OB5kbv7ordmDQ";
    private static final String PUBLIC_KEY_B64URL = "iojj3XQJ8ZX9UtstPLpdcspnCb8dlBIb83SIAbQPb1w";

    private static byte[] publicKey() {
        return Base64.getUrlDecoder().decode(PUBLIC_KEY_B64URL);
    }

    private static DiscoveryDocument document() {
        return new DiscoveryDocument(
                "https://auth.example.com", "https://registry.example.com", "https://idp.example.com",
                "https://idp.example.com/jwks.json", List.of("private_key_jwt"), List.of("x509"),
                "30d", "1", "2026-01-02T03:04:05Z", "2026-01-02T04:04:05Z", "kid-vector", "EdDSA", SIGNATURE);
    }

    @Test
    void canonicalMatchesCrossLanguageVector() {
        assertEquals(CANONICAL, new String(DiscoveryVerification.canonicalDiscovery(document()), StandardCharsets.UTF_8));
    }

    @Test
    void canonicalNormalizesFractionalSeconds() {
        DiscoveryDocument doc = new DiscoveryDocument(
                "https://auth.example.com", "https://registry.example.com", "https://idp.example.com",
                "https://idp.example.com/jwks.json", List.of("private_key_jwt"), List.of("x509"),
                "30d", "1", "2026-01-02T03:04:05.123456Z", "2026-01-02T04:04:05Z", "kid-vector", "EdDSA", SIGNATURE);
        String canonical = new String(DiscoveryVerification.canonicalDiscovery(doc), StandardCharsets.UTF_8);
        assertEquals(true, canonical.contains("\"issuedAt\":\"2026-01-02T03:04:05Z\""));
    }

    @Test
    void verifyAcceptsSignedDocument() {
        DiscoveryVerification.verifyDiscovery(document(), publicKey());
    }

    @Test
    void verifyRejectsTamperWrongKeyAndUnsigned() {
        DiscoveryDocument tampered = new DiscoveryDocument(
                "https://auth.example.com", "https://evil.example.com", "https://idp.example.com",
                "https://idp.example.com/jwks.json", List.of("private_key_jwt"), List.of("x509"),
                "30d", "1", "2026-01-02T03:04:05Z", "2026-01-02T04:04:05Z", "kid-vector", "EdDSA", SIGNATURE);
        assertThrows(DiscoveryVerificationException.class,
                () -> DiscoveryVerification.verifyDiscovery(tampered, publicKey()));

        assertThrows(DiscoveryVerificationException.class,
                () -> DiscoveryVerification.verifyDiscovery(document(), new byte[32]));

        DiscoveryDocument unsigned = new DiscoveryDocument(
                "https://auth.example.com", "https://registry.example.com", "https://idp.example.com",
                "https://idp.example.com/jwks.json", List.of("private_key_jwt"), List.of("x509"),
                "30d", "1", "2026-01-02T03:04:05Z", "2026-01-02T04:04:05Z", "kid-vector", "EdDSA", null);
        assertThrows(DiscoveryVerificationException.class,
                () -> DiscoveryVerification.verifyDiscovery(unsigned, publicKey()));
    }

    @Test
    void verifyWithJwks() {
        Map<String, Object> jwk = Map.<String, Object>of(
                "kty", "OKP", "crv", "Ed25519", "kid", "kid-vector", "alg", "EdDSA", "use", "sig",
                "x", PUBLIC_KEY_B64URL);
        Jwks jwks = new Jwks(List.of(jwk));
        DiscoveryVerification.verifyDiscoveryWithJwks(document(), jwks);
        assertArrayEquals(publicKey(), DiscoveryVerification.jwkPublicKey(jwk));

        DiscoveryDocument unknown = new DiscoveryDocument(
                "https://auth.example.com", "https://registry.example.com", "https://idp.example.com",
                "https://idp.example.com/jwks.json", List.of("private_key_jwt"), List.of("x509"),
                "30d", "1", "2026-01-02T03:04:05Z", "2026-01-02T04:04:05Z", "kid-2", "EdDSA", SIGNATURE);
        assertThrows(DiscoveryVerificationException.class,
                () -> DiscoveryVerification.verifyDiscoveryWithJwks(unknown, jwks));
    }
}
