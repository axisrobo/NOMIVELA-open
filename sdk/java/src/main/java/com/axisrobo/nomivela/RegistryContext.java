package com.axisrobo.nomivela;

/**
 * A consistent point-in-time Registry read for one Agent identity. Namespace,
 * Agent, and Identity are always present; the workload and instance are present
 * when selected.
 */
public record RegistryContext(
        Namespace namespace,
        Agent agent,
        AgentIdentity identity,
        WorkloadRegistration workloadRegistration,
        AgentInstance instance) {
}
