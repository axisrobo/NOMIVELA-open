package com.axisrobo.nomivela;

import java.util.List;

/** A business Agent Record. */
public record Agent(
        String agentRef,
        String name,
        String purpose,
        String sponsorRef,
        String ownerRef,
        String riskClass,
        String agentClass,
        List<String> carrierRefs,
        String state,
        Long agentEpoch) {
}
