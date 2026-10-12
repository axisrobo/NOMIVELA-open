package com.axisrobo.nomivela;

/** A discovery document failed verification. */
public class DiscoveryVerificationException extends IllegalArgumentException {

    public DiscoveryVerificationException(String message) {
        super(message);
    }
}
