package com.axisrobo.nomivela;

import java.util.Map;

/**
 * A record on the recoverable change stream. {@code cursor} is a global position;
 * {@code payloadVersion} identifies the minimal payload schema.
 */
public record OutboxEvent(
        String eventId,
        String eventType,
        String aggregateType,
        String aggregateId,
        Long sequence,
        Long cursor,
        Long payloadVersion,
        Map<String, Object> payload,
        String occurredAt,
        Long attempts) {
}
