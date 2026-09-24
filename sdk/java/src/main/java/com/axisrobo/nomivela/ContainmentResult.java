package com.axisrobo.nomivela;

import java.util.List;

/** The result of a cross-namespace Agent containment. */
public record ContainmentResult(
        String agentRef,
        String mode,
        String agentState,
        Long agentEpoch,
        List<String> identities,
        List<String> instances) {
}
