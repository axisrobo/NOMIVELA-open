package com.axisrobo.nomivela;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.security.SecureRandom;
import java.time.Duration;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * A zero-dependency client for the NOMIVELA Agent Registry API. Every mutation
 * sends {@code Idempotency-Key} and {@code X-Actor}; non-2xx responses raise
 * {@link ApiException} with the stable contract code.
 */
public final class Client {

    private static final SecureRandom RANDOM = new SecureRandom();

    private final String baseUrl;
    private final String actor;
    private final HttpClient http;

    public Client(String baseUrl) {
        this(baseUrl, "nomivela-sdk");
    }

    public Client(String baseUrl, String actor) {
        this.baseUrl = baseUrl.endsWith("/") ? baseUrl.substring(0, baseUrl.length() - 1) : baseUrl;
        this.actor = actor;
        this.http = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(30)).build();
    }

    // -- Namespaces ------------------------------------------------------

    public Namespace createNamespace(String namespace, String authorityRootRef) {
        return createNamespace(namespace, authorityRootRef, Mutation.none());
    }

    public Namespace createNamespace(String namespace, String authorityRootRef, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("authorityRootRef", authorityRootRef);
        mutation.applyTo(body);
        return toNamespace(request("POST", "/v1/namespaces", body));
    }

    public List<Namespace> listNamespaces() {
        return items("/v1/namespaces").stream().map(Client::toNamespace).toList();
    }

    public Namespace transitionNamespace(String namespace, String state, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("state", state);
        mutation.applyTo(body);
        return toNamespace(request("POST", "/v1/namespaces/lifecycle", body));
    }

    // -- Agents ----------------------------------------------------------

    public Agent createAgent(Agent agent, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("agentRef", agent.agentRef());
        body.put("name", agent.name());
        body.put("purpose", agent.purpose());
        body.put("sponsorRef", agent.sponsorRef());
        body.put("ownerRef", agent.ownerRef());
        body.put("riskClass", agent.riskClass());
        if (agent.agentClass() != null && !agent.agentClass().isEmpty()) {
            body.put("agentClass", agent.agentClass());
        }
        if (agent.carrierRefs() != null && !agent.carrierRefs().isEmpty()) {
            body.put("carrierRefs", agent.carrierRefs());
        }
        mutation.applyTo(body);
        return toAgent(request("POST", "/v1/agents", body));
    }

    public List<Agent> listAgents() {
        return items("/v1/agents").stream().map(Client::toAgent).toList();
    }

    public Agent transitionAgent(String agentRef, String state, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("state", state);
        mutation.applyTo(body);
        return toAgent(request("POST", "/v1/agents/" + agentRef + "/lifecycle", body));
    }

    public ContainmentResult containAgent(String agentRef, String mode, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("mode", mode);
        mutation.applyTo(body);
        return toContainment(request("POST", "/v1/agents/" + agentRef + "/containment", body));
    }

    // -- Identities ------------------------------------------------------

    public AgentIdentity allocateIdentity(String namespace, String agentRef, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("agentRef", agentRef);
        mutation.applyTo(body);
        return toIdentity(request("POST", "/v1/agent-identities", body));
    }

    public List<AgentIdentity> listIdentities(String namespace) {
        return items("/v1/agent-identities?namespace=" + encode(namespace)).stream().map(Client::toIdentity).toList();
    }

    public AgentIdentity bindIdentity(String namespace, String agentId, String rootRef, String rootType, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("authorityRootRef", rootRef);
        body.put("authorityRootType", rootType);
        mutation.applyTo(body);
        return toIdentity(request("POST", "/v1/agent-identities/" + agentId + "/binding", body));
    }

    public AgentIdentity transitionIdentity(String namespace, String agentId, String state, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("state", state);
        mutation.applyTo(body);
        return toIdentity(request("POST", "/v1/agent-identities/" + agentId + "/lifecycle", body));
    }

    // -- Workloads -------------------------------------------------------

    public WorkloadRegistration createWorkloadRegistration(String namespace, String platform, Map<String, String> selector,
                                                           String trustDomain, List<String> proofMethods, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("namespace", namespace);
        body.put("platform", platform);
        body.put("selector", selector);
        body.put("trustDomain", trustDomain);
        body.put("allowedProofMethods", proofMethods);
        mutation.applyTo(body);
        return toWorkload(request("POST", "/v1/workload-registrations", body));
    }

    public List<WorkloadRegistration> listWorkloadRegistrations(String namespace) {
        return items("/v1/workload-registrations?namespace=" + encode(namespace)).stream().map(Client::toWorkload).toList();
    }

    public WorkloadRegistration transitionWorkloadRegistration(String workloadRegistrationId, String state, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("state", state);
        mutation.applyTo(body);
        return toWorkload(request("POST", "/v1/workload-registrations/" + workloadRegistrationId + "/lifecycle", body));
    }

    // -- Instances -------------------------------------------------------

    public AgentInstance commitInstance(String agentId, InstanceCommit commit, Mutation mutation) {
        Map<String, Object> body = commit.payload();
        mutation.applyTo(body);
        return toInstance(request("POST", "/v1/agent-identities/" + agentId + "/instances", body));
    }

    public List<AgentInstance> listInstances(String namespace, String agentId) {
        return items("/v1/agent-identities/" + agentId + "/instances?namespace=" + encode(namespace)).stream().map(Client::toInstance).toList();
    }

    public AgentInstance transitionInstance(String instanceId, String state, Mutation mutation) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("state", state);
        mutation.applyTo(body);
        return toInstance(request("POST", "/v1/instances/" + instanceId + "/lifecycle", body));
    }

    // -- Evidence --------------------------------------------------------

    public List<LifecycleEvent> listEvidence() {
        return items("/v1/evidence").stream().map(Client::toLifecycleEvent).toList();
    }

    public List<OutboxEvent> listEvents() {
        return items("/v1/events").stream().map(Client::toOutboxEvent).toList();
    }

    // -- Transport -------------------------------------------------------

    private List<Map<String, Object>> items(String path) {
        Map<String, Object> payload = request("GET", path, null);
        Object rawItems = payload.get("items");
        List<Map<String, Object>> result = new ArrayList<>();
        if (rawItems instanceof List<?> list) {
            for (Object item : list) {
                if (item instanceof Map<?, ?> map) {
                    result.add(castMap(map));
                }
            }
        }
        return result;
    }

    private Map<String, Object> request(String method, String path, Map<String, Object> body) {
        HttpRequest.Builder builder = HttpRequest.newBuilder(URI.create(baseUrl + path))
                .timeout(Duration.ofSeconds(30));
        if (body == null) {
            builder.method(method, HttpRequest.BodyPublishers.noBody());
        } else {
            builder.header("Content-Type", "application/json");
            builder.method(method, HttpRequest.BodyPublishers.ofString(Json.write(body)));
        }
        if (!method.equals("GET")) {
            builder.header("Idempotency-Key", newIdempotencyKey());
            builder.header("X-Actor", actor);
        }

        HttpResponse<String> response;
        try {
            response = http.send(builder.build(), HttpResponse.BodyHandlers.ofString());
        } catch (IOException error) {
            throw new IllegalStateException("nomivela: request failed", error);
        } catch (InterruptedException error) {
            Thread.currentThread().interrupt();
            throw new IllegalStateException("nomivela: interrupted", error);
        }

        if (response.statusCode() >= 300) {
            throw apiError(response);
        }
        if (response.body() == null || response.body().isEmpty()) {
            return Map.of();
        }
        return Json.parseObject(response.body());
    }

    private static ApiException apiError(HttpResponse<String> response) {
        String code = "";
        String message = response.body();
        String correlation = "";
        try {
            Map<String, Object> envelope = Json.parseObject(response.body());
            if (envelope.get("error") instanceof Map<?, ?> detail) {
                Map<String, Object> error = castMap(detail);
                code = text(error, "code");
                message = text(error, "message");
                correlation = text(error, "correlationId");
            }
        } catch (RuntimeException ignored) {
            // fall back to the raw body
        }
        return new ApiException(response.statusCode(), code, message, correlation);
    }

    private static String newIdempotencyKey() {
        byte[] bytes = new byte[16];
        RANDOM.nextBytes(bytes);
        return HexFormat.of().formatHex(bytes);
    }

    private static String encode(String value) {
        return java.net.URLEncoder.encode(value, java.nio.charset.StandardCharsets.UTF_8);
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> castMap(Map<?, ?> map) {
        return (Map<String, Object>) map;
    }

    private static String text(Map<String, Object> map, String key) {
        Object value = map.get(key);
        return value == null ? "" : value.toString();
    }

    private static Long number(Map<String, Object> map, String key) {
        Object value = map.get(key);
        return value instanceof Number number ? number.longValue() : null;
    }

    private static List<String> strings(Map<String, Object> map, String key) {
        Object value = map.get(key);
        if (!(value instanceof List<?> list)) {
            return null;
        }
        List<String> result = new ArrayList<>();
        for (Object item : list) {
            result.add(String.valueOf(item));
        }
        return result;
    }

    private static Map<String, String> stringMap(Map<String, Object> map, String key) {
        Object value = map.get(key);
        if (!(value instanceof Map<?, ?> nested)) {
            return null;
        }
        Map<String, String> result = new LinkedHashMap<>();
        for (Map.Entry<?, ?> entry : nested.entrySet()) {
            result.put(String.valueOf(entry.getKey()), String.valueOf(entry.getValue()));
        }
        return result;
    }

    private static Namespace toNamespace(Map<String, Object> map) {
        return new Namespace(text(map, "namespace"), text(map, "parent"), text(map, "authorityRootRef"),
                text(map, "status"), number(map, "namespaceEpoch"));
    }

    private static Agent toAgent(Map<String, Object> map) {
        return new Agent(text(map, "agentRef"), text(map, "name"), text(map, "purpose"), text(map, "sponsorRef"),
                text(map, "ownerRef"), text(map, "riskClass"), text(map, "agentClass"), strings(map, "carrierRefs"),
                text(map, "state"), number(map, "agentEpoch"));
    }

    private static AgentIdentity toIdentity(Map<String, Object> map) {
        return new AgentIdentity(text(map, "namespace"), text(map, "agentId"), text(map, "agentRef"),
                text(map, "state"), number(map, "identityEpoch"), text(map, "authorityRootRef"),
                text(map, "authorityRootType"));
    }

    private static WorkloadRegistration toWorkload(Map<String, Object> map) {
        return new WorkloadRegistration(text(map, "workloadRegistrationId"), text(map, "namespace"),
                text(map, "platform"), stringMap(map, "selector"), text(map, "trustDomain"),
                strings(map, "allowedProofMethods"), text(map, "status"), number(map, "workloadEpoch"));
    }

    private static AgentInstance toInstance(Map<String, Object> map) {
        return new AgentInstance(text(map, "instanceId"), text(map, "namespace"), text(map, "agentId"),
                text(map, "workloadRegistrationId"), text(map, "workloadId"), text(map, "artifactDigest"),
                text(map, "attestationRef"), text(map, "leaseExpiresAt"), text(map, "state"), number(map, "generation"));
    }

    private static LifecycleEvent toLifecycleEvent(Map<String, Object> map) {
        return new LifecycleEvent(text(map, "eventId"), text(map, "objectType"), text(map, "objectId"),
                text(map, "previousState"), text(map, "newState"), text(map, "epochKind"), number(map, "epoch"),
                text(map, "actor"), text(map, "reason"), text(map, "evidenceRef"), text(map, "occurredAt"));
    }

    private static OutboxEvent toOutboxEvent(Map<String, Object> map) {
        return new OutboxEvent(text(map, "eventId"), text(map, "eventType"), text(map, "aggregateType"),
                text(map, "aggregateId"), number(map, "sequence"), text(map, "occurredAt"));
    }

    private static ContainmentResult toContainment(Map<String, Object> map) {
        return new ContainmentResult(text(map, "agentRef"), text(map, "mode"), text(map, "agentState"),
                number(map, "agentEpoch"), strings(map, "identities"), strings(map, "instances"));
    }
}
