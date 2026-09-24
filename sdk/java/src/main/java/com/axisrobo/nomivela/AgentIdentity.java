package com.axisrobo.nomivela;

/** A security Agent Identity Record. */
public record AgentIdentity(
        String namespace,
        String agentId,
        String agentRef,
        String state,
        Long identityEpoch,
        String authorityRootRef,
        String authorityRootType) {
}
