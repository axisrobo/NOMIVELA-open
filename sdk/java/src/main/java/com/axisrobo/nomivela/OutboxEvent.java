package com.axisrobo.nomivela;

/** A transactional outbox record. */
public record OutboxEvent(
        String eventId,
        String eventType,
        String aggregateType,
        String aggregateId,
        Long sequence,
        String occurredAt) {
}
