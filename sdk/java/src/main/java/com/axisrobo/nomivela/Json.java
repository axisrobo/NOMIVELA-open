package com.axisrobo.nomivela;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * A minimal JSON reader and writer so the SDK has no runtime dependencies.
 */
final class Json {

    private final String source;
    private int position;

    private Json(String source) {
        this.source = source;
    }

    static Object parse(String text) {
        Json parser = new Json(text);
        parser.skipWhitespace();
        Object value = parser.readValue();
        parser.skipWhitespace();
        if (parser.position != parser.source.length()) {
            throw new IllegalArgumentException("unexpected trailing content in JSON");
        }
        return value;
    }

    @SuppressWarnings("unchecked")
    static Map<String, Object> parseObject(String text) {
        Object value = parse(text);
        if (!(value instanceof Map)) {
            throw new IllegalArgumentException("expected a JSON object");
        }
        return (Map<String, Object>) value;
    }

    private Object readValue() {
        char current = peek();
        switch (current) {
            case '{':
                return readObject();
            case '[':
                return readArray();
            case '"':
                return readString();
            case 't':
                expect("true");
                return Boolean.TRUE;
            case 'f':
                expect("false");
                return Boolean.FALSE;
            case 'n':
                expect("null");
                return null;
            default:
                return readNumber();
        }
    }

    private Map<String, Object> readObject() {
        Map<String, Object> object = new LinkedHashMap<>();
        position++; // {
        skipWhitespace();
        if (peek() == '}') {
            position++;
            return object;
        }
        while (true) {
            skipWhitespace();
            String key = readString();
            skipWhitespace();
            if (peek() != ':') {
                throw new IllegalArgumentException("expected ':' in JSON object");
            }
            position++;
            skipWhitespace();
            object.put(key, readValue());
            skipWhitespace();
            char next = peek();
            position++;
            if (next == '}') {
                return object;
            }
            if (next != ',') {
                throw new IllegalArgumentException("expected ',' or '}' in JSON object");
            }
        }
    }

    private List<Object> readArray() {
        List<Object> array = new ArrayList<>();
        position++; // [
        skipWhitespace();
        if (peek() == ']') {
            position++;
            return array;
        }
        while (true) {
            skipWhitespace();
            array.add(readValue());
            skipWhitespace();
            char next = peek();
            position++;
            if (next == ']') {
                return array;
            }
            if (next != ',') {
                throw new IllegalArgumentException("expected ',' or ']' in JSON array");
            }
        }
    }

    private String readString() {
        if (peek() != '"') {
            throw new IllegalArgumentException("expected a JSON string");
        }
        position++;
        StringBuilder builder = new StringBuilder();
        while (true) {
            char current = source.charAt(position++);
            if (current == '"') {
                return builder.toString();
            }
            if (current == '\\') {
                char escape = source.charAt(position++);
                switch (escape) {
                    case '"' -> builder.append('"');
                    case '\\' -> builder.append('\\');
                    case '/' -> builder.append('/');
                    case 'b' -> builder.append('\b');
                    case 'f' -> builder.append('\f');
                    case 'n' -> builder.append('\n');
                    case 'r' -> builder.append('\r');
                    case 't' -> builder.append('\t');
                    case 'u' -> {
                        builder.append((char) Integer.parseInt(source.substring(position, position + 4), 16));
                        position += 4;
                    }
                    default -> throw new IllegalArgumentException("invalid JSON escape");
                }
                continue;
            }
            builder.append(current);
        }
    }

    private Object readNumber() {
        int start = position;
        while (position < source.length() && "-+.eE0123456789".indexOf(source.charAt(position)) >= 0) {
            position++;
        }
        if (start == position) {
            throw new IllegalArgumentException("invalid JSON value");
        }
        return Double.valueOf(source.substring(start, position));
    }

    private void expect(String literal) {
        if (!source.startsWith(literal, position)) {
            throw new IllegalArgumentException("invalid JSON literal");
        }
        position += literal.length();
    }

    private char peek() {
        if (position >= source.length()) {
            throw new IllegalArgumentException("unexpected end of JSON");
        }
        return source.charAt(position);
    }

    private void skipWhitespace() {
        while (position < source.length() && Character.isWhitespace(source.charAt(position))) {
            position++;
        }
    }

    static String write(Object value) {
        StringBuilder builder = new StringBuilder();
        writeValue(builder, value);
        return builder.toString();
    }

    private static void writeValue(StringBuilder builder, Object value) {
        if (value == null) {
            builder.append("null");
        } else if (value instanceof String text) {
            writeString(builder, text);
        } else if (value instanceof Boolean || value instanceof Number) {
            builder.append(value);
        } else if (value instanceof Map<?, ?> map) {
            builder.append('{');
            boolean first = true;
            for (Map.Entry<?, ?> entry : map.entrySet()) {
                if (!first) {
                    builder.append(',');
                }
                first = false;
                writeString(builder, String.valueOf(entry.getKey()));
                builder.append(':');
                writeValue(builder, entry.getValue());
            }
            builder.append('}');
        } else if (value instanceof Iterable<?> iterable) {
            builder.append('[');
            boolean first = true;
            for (Object item : iterable) {
                if (!first) {
                    builder.append(',');
                }
                first = false;
                writeValue(builder, item);
            }
            builder.append(']');
        } else {
            writeString(builder, value.toString());
        }
    }

    private static void writeString(StringBuilder builder, String text) {
        builder.append('"');
        for (int i = 0; i < text.length(); i++) {
            char current = text.charAt(i);
            switch (current) {
                case '"' -> builder.append("\\\"");
                case '\\' -> builder.append("\\\\");
                case '\n' -> builder.append("\\n");
                case '\r' -> builder.append("\\r");
                case '\t' -> builder.append("\\t");
                default -> {
                    if (current < 0x20) {
                        builder.append(String.format("\\u%04x", (int) current));
                    } else {
                        builder.append(current);
                    }
                }
            }
        }
        builder.append('"');
    }
}
