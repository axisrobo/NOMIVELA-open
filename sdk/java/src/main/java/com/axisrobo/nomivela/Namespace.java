package com.axisrobo.nomivela;

/** An authority namespace. */
public record Namespace(
        String namespace,
        String parent,
        String authorityRootRef,
        String status,
        Long namespaceEpoch) {
}
