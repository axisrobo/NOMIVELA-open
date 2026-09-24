package com.axisrobo.nomivela;

/** Append-only evidence of a state transition. */
public record LifecycleEvent(
        String eventId,
        String objectType,
        String objectId,
        String previousState,
        String newState,
        String epochKind,
        Long epoch,
        String actor,
        String reason,
        String evidenceRef,
        String occurredAt) {
}
