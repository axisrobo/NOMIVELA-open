import json
import threading
import unittest
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

import nomivela
from nomivela import Agent, APIError, Client, InstanceCommit, Mutation


class StubServer:
    """A minimal HTTP stub that records requests and returns canned responses."""

    def __init__(self, respond):
        self.respond = respond
        self.requests = []
        outer = self

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, *args):  # silence the default logging
                pass

            def _handle(self):
                length = int(self.headers.get("Content-Length") or 0)
                raw = self.rfile.read(length) if length else b""
                body = json.loads(raw) if raw else None
                parsed = urlparse(self.path)
                outer.requests.append(
                    {
                        "method": self.command,
                        "path": parsed.path,
                        "query": parse_qs(parsed.query),
                        "headers": dict(self.headers),
                        "body": body,
                    }
                )
                status, payload = outer.respond(self.command, parsed.path, body)
                data = json.dumps(payload).encode("utf-8") if payload is not None else b""
                self.send_response(status)
                if data:
                    self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                if data:
                    self.wfile.write(data)

            do_GET = _handle
            do_POST = _handle
            do_PUT = _handle

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    @property
    def url(self) -> str:
        return f"http://127.0.0.1:{self.server.server_address[1]}"

    def close(self):
        self.server.shutdown()
        self.server.server_close()


class ClientTest(unittest.TestCase):
    def setUp(self):
        self.stub = None

    def tearDown(self):
        if self.stub:
            self.stub.close()

    def test_create_namespace_sends_contract_headers_and_body(self):
        def respond(method, path, body):
            return 201, {
                "namespace": "https://auth.example.com",
                "authorityRootRef": "root:org-a",
                "status": "active",
                "namespaceEpoch": 1,
            }

        self.stub = StubServer(respond)
        client = Client(self.stub.url, actor="user:ci")
        namespace = client.create_namespace(
            "https://auth.example.com", "root:org-a", Mutation(reason="bootstrap")
        )

        self.assertEqual(namespace.status, "active")
        self.assertEqual(namespace.namespace_epoch, 1)

        request = self.stub.requests[0]
        self.assertEqual(request["path"], "/v1/namespaces")
        self.assertTrue(request["headers"].get("Idempotency-Key"))
        self.assertEqual(request["headers"].get("X-Actor"), "user:ci")
        self.assertEqual(request["body"]["namespace"], "https://auth.example.com")
        self.assertEqual(request["body"]["reason"], "bootstrap")

    def test_list_identities_encodes_query(self):
        def respond(method, path, body):
            return 200, {"items": [{"agentId": "abc", "state": "bound", "identityEpoch": 2}]}

        self.stub = StubServer(respond)
        identities = Client(self.stub.url).list_identities("https://auth.example.com")

        self.assertEqual(len(identities), 1)
        self.assertEqual(identities[0].agent_id, "abc")
        self.assertEqual(identities[0].identity_epoch, 2)
        self.assertEqual(
            self.stub.requests[0]["query"]["namespace"], ["https://auth.example.com"]
        )

    def test_api_error_carries_code(self):
        def respond(method, path, body):
            return 409, {
                "error": {
                    "code": "conflict",
                    "message": "an active agent id already exists",
                    "correlationId": "corr-9",
                }
            }

        self.stub = StubServer(respond)
        with self.assertRaises(APIError) as raised:
            Client(self.stub.url).create_namespace("https://auth.example.com", "root:org-a")

        error = raised.exception
        self.assertEqual(error.status, 409)
        self.assertEqual(error.code, "conflict")
        self.assertEqual(error.correlation_id, "corr-9")

    def test_contain_agent_and_commit_instance(self):
        def respond(method, path, body):
            if path == "/v1/agents/agent_order_processor/containment":
                return 200, {
                    "agentRef": "agent_order_processor",
                    "mode": "quarantine",
                    "agentState": "suspended",
                    "agentEpoch": 5,
                    "identities": ["id-1"],
                    "instances": ["inst-1"],
                }
            if path == "/v1/agent-identities/id-1/instances":
                return 201, {"instanceId": "inst-1", "state": "active", "generation": 1}
            raise AssertionError(f"unexpected path {path}")

        self.stub = StubServer(respond)
        client = Client(self.stub.url)

        containment = client.contain_agent("agent_order_processor", "quarantine", Mutation(reason="incident"))
        self.assertEqual(containment.agent_state, "suspended")
        self.assertEqual(containment.instances, ["inst-1"])

        instance = client.commit_instance(
            "id-1",
            InstanceCommit(
                namespace="https://auth.example.com",
                workload_registration_id="wr-1",
                workload_id="workload-1",
                artifact_digest="sha256:abc",
                attestation_ref="attestation:1",
                lease_expires_at=datetime(2999, 1, 1, tzinfo=timezone.utc),
            ),
        )
        self.assertEqual(instance.state, "active")
        commit_body = self.stub.requests[1]["body"]
        self.assertEqual(commit_body["leaseExpiresAt"], "2999-01-01T00:00:00Z")

    def test_create_agent_sends_form_and_carriers(self):
        def respond(method, path, body):
            return 201, {"agentRef": "agent_edge_twin", "agentClass": "asset_twin", "carrierRefs": ["device:gateway-7"], "state": "draft", "agentEpoch": 1}

        self.stub = StubServer(respond)
        agent = Client(self.stub.url).create_agent(
            Agent(
                agent_ref="agent_edge_twin",
                name="Edge Twin",
                purpose="Mirror the gateway",
                sponsor_ref="org:platform",
                owner_ref="user:owner",
                risk_class="high",
                agent_class="asset_twin",
                carrier_refs=["device:gateway-7"],
            )
        )

        self.assertEqual(agent.agent_class, "asset_twin")
        body = self.stub.requests[0]["body"]
        self.assertEqual(body["agentClass"], "asset_twin")
        self.assertEqual(body["carrierRefs"], ["device:gateway-7"])
        self.assertNotIn("state", body)


if __name__ == "__main__":
    unittest.main()
