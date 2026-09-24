package com.axisrobo.nomivela;

import java.util.List;
import java.util.Map;

/** An approved workload declaration. */
public record WorkloadRegistration(
        String workloadRegistrationId,
        String namespace,
        String platform,
        Map<String, String> selector,
        String trustDomain,
        List<String> allowedProofMethods,
        String status,
        Long workloadEpoch) {
}
