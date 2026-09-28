package com.axisrobo.nomivela;

import java.util.List;

/**
 * The registry discovery document. A signed deployment adds
 * {@code discoveryVersion}, {@code issuedAt}, {@code expiresAt},
 * {@code signingKid}, {@code alg}, and {@code signature}.
 */
public record DiscoveryDocument(
        String namespace,
        String registryEndpoint,
        String issuer,
        String jwksUri,
        List<String> supportedProofProfiles,
        List<String> supportedArtifactTypes,
        String keyRotation,
        String discoveryVersion,
        String issuedAt,
        String expiresAt,
        String signingKid,
        String alg,
        String signature) {
}
