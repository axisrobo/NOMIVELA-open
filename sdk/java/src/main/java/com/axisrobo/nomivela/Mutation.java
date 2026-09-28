package com.axisrobo.nomivela;

import java.util.Map;

/**
 * Attribution recorded for a state change.
 *
 * <p>{@code expectedEpoch} is an optional optimistic-concurrency precondition:
 * when non-zero the object's current epoch must match or the mutation conflicts.
 */
public record Mutation(String reason, String evidenceRef, long expectedEpoch) {

    public Mutation(String reason, String evidenceRef) {
        this(reason, evidenceRef, 0L);
    }

    public static Mutation none() {
        return new Mutation(null, null, 0L);
    }

    void applyTo(Map<String, Object> body) {
        if (reason != null && !reason.isEmpty()) {
            body.put("reason", reason);
        }
        if (evidenceRef != null && !evidenceRef.isEmpty()) {
            body.put("evidenceRef", evidenceRef);
        }
        if (expectedEpoch > 0) {
            body.put("expectedEpoch", expectedEpoch);
        }
    }
}
