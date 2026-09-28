package com.axisrobo.nomivela;

import java.util.List;

/** A page of the change stream. */
public record EventPage(List<OutboxEvent> items, Long nextCursor) {
}
