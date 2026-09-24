package com.axisrobo.nomivela;

/** An error response from the API, carrying the stable contract code. */
public class ApiException extends RuntimeException {

    private final int status;
    private final String code;
    private final String correlationId;

    public ApiException(int status, String code, String message, String correlationId) {
        super("nomivela: " + message + " (status " + status + ", code " + code + ", correlation " + correlationId + ")");
        this.status = status;
        this.code = code;
        this.correlationId = correlationId;
    }

    public int status() {
        return status;
    }

    public String code() {
        return code;
    }

    public String correlationId() {
        return correlationId;
    }
}
