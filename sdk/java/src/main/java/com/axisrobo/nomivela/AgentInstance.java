package com.axisrobo.nomivela;

/** A running instance of an Agent identity. */
public record AgentInstance(
        String instanceId,
        String namespace,
        String agentId,
        String workloadRegistrationId,
        String workloadId,
        String artifactDigest,
        String attestationRef,
        String leaseExpiresAt,
        String state,
        Long generation) {
}
