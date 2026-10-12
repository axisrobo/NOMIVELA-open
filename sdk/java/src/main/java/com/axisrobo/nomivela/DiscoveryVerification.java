package com.axisrobo.nomivela;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.KeyFactory;
import java.security.PublicKey;
import java.security.Signature;
import java.security.spec.X509EncodedKeySpec;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.time.temporal.ChronoUnit;
import java.util.Base64;
import java.util.Map;
import java.util.TreeMap;

/**
 * Verifies a signed discovery document using only the public contract — no Core
 * internals. A caller fetches the document and verifies its signature against a
 * key it trusts independently. Verifying a signature does not establish
 * freshness; the caller must also enforce the validity window and its
 * namespace/endpoint policy.
 */
public final class DiscoveryVerification {

    private static final DateTimeFormatter RFC3339_SECONDS =
            DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm:ss'Z'");

    // DER prefix of the Ed25519 SubjectPublicKeyInfo; JCA needs a wrapped key.
    private static final byte[] ED25519_SPKI_PREFIX = {
        0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00
    };

    private DiscoveryVerification() {
    }

    /** The deterministic bytes the registry signs: keys sorted, HTML escaping
     * disabled, second-precision RFC 3339 UTC timestamps. */
    public static byte[] canonicalDiscovery(DiscoveryDocument document) {
        Map<String, Object> payload = new TreeMap<>();
        payload.put("alg", orEmpty(document.alg()));
        payload.put("discoveryVersion", orEmpty(document.discoveryVersion()));
        payload.put("expiresAt", canonicalTime(document.expiresAt()));
        payload.put("issuedAt", canonicalTime(document.issuedAt()));
        payload.put("issuer", orEmpty(document.issuer()));
        payload.put("jwksUri", orEmpty(document.jwksUri()));
        payload.put("keyRotation", orEmpty(document.keyRotation()));
        payload.put("namespace", orEmpty(document.namespace()));
        payload.put("registryEndpoint", orEmpty(document.registryEndpoint()));
        payload.put("signingKid", orEmpty(document.signingKid()));
        payload.put("supportedArtifactTypes", document.supportedArtifactTypes());
        payload.put("supportedProofProfiles", document.supportedProofProfiles());
        return Json.write(payload).getBytes(StandardCharsets.UTF_8);
    }

    /** Verify {@code document.signature()} against {@code publicKey} (raw 32
     * bytes). Does not enforce the validity window or trust the key. */
    public static void verifyDiscovery(DiscoveryDocument document, byte[] publicKey) {
        if (document.signature() == null || document.signature().isEmpty()) {
            throw new DiscoveryVerificationException("discovery document is not signed");
        }
        if (publicKey == null || publicKey.length != 32) {
            throw new DiscoveryVerificationException("verification key must be 32 bytes");
        }
        byte[] signature = decodeBase64Url(document.signature());
        byte[] canonical = canonicalDiscovery(document);
        try {
            Signature verifier = Signature.getInstance("Ed25519");
            verifier.initVerify(publicKey(publicKey));
            verifier.update(canonical);
            if (!verifier.verify(signature)) {
                throw new DiscoveryVerificationException("discovery signature does not verify");
            }
        } catch (GeneralSecurityException error) {
            throw new DiscoveryVerificationException(error.getMessage());
        }
    }

    /** Decode an Ed25519 verification JWK into a raw 32-byte public key. */
    public static byte[] jwkPublicKey(Map<String, Object> jwk) {
        if (!"OKP".equals(jwk.get("kty")) || !"Ed25519".equals(jwk.get("crv"))) {
            throw new DiscoveryVerificationException(
                    "unsupported JWK kty=" + jwk.get("kty") + " crv=" + jwk.get("crv"));
        }
        byte[] raw = decodeBase64Url(String.valueOf(jwk.get("x")));
        if (raw.length != 32) {
            throw new DiscoveryVerificationException("JWK x must be 32 bytes, got " + raw.length);
        }
        return raw;
    }

    /** Verify using the JWKS key named by {@code signingKid}. Trust the JWKS
     * independently; fetching it from the same origin as the document does not
     * establish trust. */
    public static void verifyDiscoveryWithJwks(DiscoveryDocument document, Jwks jwks) {
        for (Map<String, Object> jwk : jwks.keys()) {
            if (!String.valueOf(document.signingKid()).equals(String.valueOf(jwk.get("kid")))) {
                continue;
            }
            verifyDiscovery(document, jwkPublicKey(jwk));
            return;
        }
        throw new DiscoveryVerificationException("no JWKS key for signing kid " + document.signingKid());
    }

    private static String canonicalTime(String value) {
        if (value == null || value.isEmpty()) {
            return "";
        }
        OffsetDateTime parsed = OffsetDateTime.parse(value);
        return parsed.withOffsetSameInstant(ZoneOffset.UTC).truncatedTo(ChronoUnit.SECONDS).format(RFC3339_SECONDS);
    }

    private static String orEmpty(String value) {
        return value == null ? "" : value;
    }

    private static byte[] decodeBase64Url(String value) {
        return Base64.getUrlDecoder().decode(value);
    }

    private static PublicKey publicKey(byte[] raw) throws GeneralSecurityException {
        byte[] der = new byte[ED25519_SPKI_PREFIX.length + raw.length];
        System.arraycopy(ED25519_SPKI_PREFIX, 0, der, 0, ED25519_SPKI_PREFIX.length);
        System.arraycopy(raw, 0, der, ED25519_SPKI_PREFIX.length, raw.length);
        return KeyFactory.getInstance("Ed25519").generatePublic(new X509EncodedKeySpec(der));
    }
}
