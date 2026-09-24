package com.axisrobo.nomivela;

import java.util.Map;

/** Attribution recorded for a state change. */
public record Mutation(String reason, String evidenceRef) {

    public static Mutation none() {
        return new Mutation(null, null);
    }

    void applyTo(Map<String, Object> body) {
        if (reason != null && !reason.isEmpty()) {
            body.put("reason", reason);
        }
        if (evidenceRef != null && !evidenceRef.isEmpty()) {
            body.put("evidenceRef", evidenceRef);
        }
    }
}
