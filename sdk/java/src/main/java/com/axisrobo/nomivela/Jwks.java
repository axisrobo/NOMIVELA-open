package com.axisrobo.nomivela;

import java.util.List;
import java.util.Map;

/** A JSON Web Key Set of registry discovery signing keys. */
public record Jwks(List<Map<String, Object>> keys) {
}
