package com.axisrobo.nomivela;

import java.util.LinkedHashMap;
import java.util.Map;

/** The verified enrollment submitted for an Agent instance. */
public record InstanceCommit(
        String namespace,
        String workloadRegistrationId,
        String workloadId,
        String artifactDigest,
        String attestationRef,
        String leaseExpiresAt) {

    Map<String, Object> payload() {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("workloadRegistrationId", workloadRegistrationId);
        body.put("workloadId", workloadId);
        body.put("artifactDigest", artifactDigest);
        body.put("attestationRef", attestationRef);
        body.put("leaseExpiresAt", leaseExpiresAt);
        return body;
    }
}
