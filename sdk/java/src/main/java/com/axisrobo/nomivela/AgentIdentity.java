package com.axisrobo.nomivela;

/** A security Agent Identity Record. */
public record AgentIdentity(
        String namespace,
        String agentId,
        String agentRef,
        String identityType,
        String subjectRef,
        String state,
        Long identityEpoch,
        String authorityRootRef,
        String authorityRootType) {

    /** Backward-compatible constructor for legacy untyped identities. */
    public AgentIdentity(
            String namespace,
            String agentId,
            String agentRef,
            String state,
            Long identityEpoch,
            String authorityRootRef,
            String authorityRootType) {
        this(namespace, agentId, agentRef, null, null, state, identityEpoch, authorityRootRef, authorityRootType);
    }
}
