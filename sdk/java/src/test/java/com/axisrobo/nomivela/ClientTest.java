package com.axisrobo.nomivela;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.function.Function;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class ClientTest {

    private record Recorded(String method, String path, String query, Map<String, List<String>> headers, String body) {
    }

    private record Stub(int status, String body) {
    }

    private HttpServer server;
    private String baseUrl;
    private final List<Recorded> recorded = Collections.synchronizedList(new ArrayList<>());
    private Function<Recorded, Stub> handler;

    @BeforeEach
    void start() throws IOException {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", this::handle);
        server.start();
        baseUrl = "http://127.0.0.1:" + server.getAddress().getPort();
    }

    @AfterEach
    void stop() {
        server.stop(0);
    }

    private static String firstHeader(Map<String, List<String>> headers, String name) {
        for (Map.Entry<String, List<String>> entry : headers.entrySet()) {
            if (entry.getKey().equalsIgnoreCase(name) && !entry.getValue().isEmpty()) {
                return entry.getValue().get(0);
            }
        }
        return null;
    }

    private void handle(HttpExchange exchange) throws IOException {
        String body = new String(exchange.getRequestBody().readAllBytes(), StandardCharsets.UTF_8);
        Recorded request = new Recorded(exchange.getRequestMethod(), exchange.getRequestURI().getPath(),
                exchange.getRequestURI().getQuery(), exchange.getRequestHeaders(), body);
        recorded.add(request);

        Stub response = handler.apply(request);
        byte[] payload = response.body() == null ? new byte[0] : response.body().getBytes(StandardCharsets.UTF_8);
        if (payload.length > 0) {
            exchange.getResponseHeaders().add("Content-Type", "application/json");
        }
        exchange.sendResponseHeaders(response.status(), payload.length);
        exchange.getResponseBody().write(payload);
        exchange.close();
    }

    @Test
    void createNamespaceSendsContractHeadersAndBody() {
        handler = request -> new Stub(201,
                "{\"namespace\":\"https://auth.example.com\",\"authorityRootRef\":\"root:org-a\",\"status\":\"active\",\"namespaceEpoch\":1}");

        Client client = new Client(baseUrl, "user:ci");
        Namespace namespace = client.createNamespace("https://auth.example.com", "root:org-a",
                new Mutation("bootstrap", null));

        assertEquals("active", namespace.status());
        assertEquals(1L, namespace.namespaceEpoch());

        Recorded request = recorded.get(0);
        assertEquals("/v1/namespaces", request.path());
        assertNotNull(firstHeader(request.headers(), "Idempotency-Key"));
        assertEquals("user:ci", firstHeader(request.headers(), "X-Actor"));

        Map<String, Object> body = Json.parseObject(request.body());
        assertEquals("https://auth.example.com", body.get("namespace"));
        assertEquals("bootstrap", body.get("reason"));
    }

    @Test
    void listIdentitiesEncodesQuery() {
        handler = request -> new Stub(200, "{\"items\":[{\"agentId\":\"abc\",\"state\":\"bound\",\"identityEpoch\":2}]}");

        List<AgentIdentity> identities = new Client(baseUrl).listIdentities("https://auth.example.com");

        assertEquals(1, identities.size());
        assertEquals("abc", identities.get(0).agentId());
        assertEquals(2L, identities.get(0).identityEpoch());

        String query = recorded.get(0).query();
        assertTrue(query.contains("namespace="));
        String decoded = URLDecoder.decode(query, StandardCharsets.UTF_8);
        assertTrue(decoded.contains("namespace=https://auth.example.com"));
    }

    @Test
    void apiErrorCarriesCode() {
        handler = request -> new Stub(409,
                "{\"error\":{\"code\":\"conflict\",\"message\":\"an active agent id already exists\",\"correlationId\":\"corr-9\"}}");

        ApiException error = assertThrows(ApiException.class,
                () -> new Client(baseUrl).createNamespace("https://auth.example.com", "root:org-a"));

        assertEquals(409, error.status());
        assertEquals("conflict", error.code());
        assertEquals("corr-9", error.correlationId());
    }

    @Test
    void containAgentAndCommitInstance() {
        handler = request -> {
            if (request.path().equals("/v1/agents/agent_order_processor/containment")) {
                return new Stub(200,
                        "{\"agentRef\":\"agent_order_processor\",\"mode\":\"quarantine\",\"agentState\":\"suspended\",\"agentEpoch\":5,\"identities\":[\"id-1\"],\"instances\":[\"inst-1\"]}");
            }
            if (request.path().equals("/v1/agent-identities/id-1/instances")) {
                return new Stub(201, "{\"instanceId\":\"inst-1\",\"state\":\"active\",\"generation\":1}");
            }
            throw new IllegalStateException("unexpected path " + request.path());
        };

        Client client = new Client(baseUrl);
        ContainmentResult containment = client.containAgent("agent_order_processor", "quarantine",
                new Mutation("incident", null));
        assertEquals("suspended", containment.agentState());
        assertEquals(List.of("inst-1"), containment.instances());

        AgentInstance instance = client.commitInstance("id-1",
                new InstanceCommit("https://auth.example.com", "wr-1", "workload-1", "sha256:abc", "attestation:1",
                        "2999-01-01T00:00:00Z"),
                Mutation.none());
        assertEquals("active", instance.state());

        Map<String, Object> commitBody = Json.parseObject(recorded.get(1).body());
        assertEquals("2999-01-01T00:00:00Z", commitBody.get("leaseExpiresAt"));
    }

    @Test
    void createAgentSendsFormAndCarriersWithoutState() {
        handler = request -> new Stub(201,
                "{\"agentRef\":\"agent_edge_twin\",\"agentClass\":\"asset_twin\",\"carrierRefs\":[\"device:gateway-7\"],\"state\":\"draft\",\"agentEpoch\":1}");

        Agent agent = new Client(baseUrl).createAgent(new Agent("agent_edge_twin", "Edge Twin", "Mirror the gateway",
                "org:platform", "user:owner", "high", "asset_twin", List.of("device:gateway-7"), null, null),
                Mutation.none());

        assertEquals("asset_twin", agent.agentClass());
        Map<String, Object> body = Json.parseObject(recorded.get(0).body());
        assertEquals("asset_twin", body.get("agentClass"));
        assertEquals(List.of("device:gateway-7"), body.get("carrierRefs"));
        assertFalse(body.containsKey("state"));
    }
}
