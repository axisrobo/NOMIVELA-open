"""Standard-library client for the NOMIVELA Agent Registry API."""

from __future__ import annotations

import json
import secrets
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from typing import Any, Optional

from .types import (
    Agent,
    AgentIdentity,
    AgentInstance,
    ContainmentResult,
    InstanceCommit,
    LifecycleEvent,
    Namespace,
    OutboxEvent,
    WorkloadRegistration,
    _build,
)


class APIError(Exception):
    """An error response from the API, carrying the stable contract code."""

    def __init__(self, status: int, code: str, message: str, correlation_id: str):
        super().__init__(
            f"nomivela: {message} (status {status}, code {code}, correlation {correlation_id})"
        )
        self.status = status
        self.code = code
        self.message = message
        self.correlation_id = correlation_id


@dataclass
class Mutation:
    """Attribution recorded for a state change."""

    reason: Optional[str] = None
    evidence_ref: Optional[str] = None

    def payload(self) -> dict[str, str]:
        out: dict[str, str] = {}
        if self.reason:
            out["reason"] = self.reason
        if self.evidence_ref:
            out["evidenceRef"] = self.evidence_ref
        return out


class Client:
    """A NOMIVELA Agent Registry API client."""

    def __init__(self, base_url: str, actor: str = "nomivela-sdk", timeout: float = 30.0):
        self.base_url = base_url.rstrip("/")
        self.actor = actor
        self.timeout = timeout

    # -- Namespaces ------------------------------------------------------

    def create_namespace(self, namespace: str, authority_root_ref: str, mutation: Optional[Mutation] = None) -> Namespace:
        body = {"namespace": namespace, "authorityRootRef": authority_root_ref}
        body.update((mutation or Mutation()).payload())
        return _build(Namespace, self._request("POST", "/v1/namespaces", body))

    def list_namespaces(self) -> list[Namespace]:
        return [_build(Namespace, item) for item in self._items("/v1/namespaces")]

    def transition_namespace(self, namespace: str, state: str, mutation: Optional[Mutation] = None) -> Namespace:
        body = {"namespace": namespace, "state": state}
        body.update((mutation or Mutation()).payload())
        return _build(Namespace, self._request("POST", "/v1/namespaces/lifecycle", body))

    # -- Agents ----------------------------------------------------------

    def create_agent(self, agent: Agent, mutation: Optional[Mutation] = None) -> Agent:
        body: dict[str, Any] = {
            "agentRef": agent.agent_ref,
            "name": agent.name,
            "purpose": agent.purpose,
            "sponsorRef": agent.sponsor_ref,
            "ownerRef": agent.owner_ref,
            "riskClass": agent.risk_class,
        }
        if agent.agent_class:
            body["agentClass"] = agent.agent_class
        if agent.carrier_refs:
            body["carrierRefs"] = agent.carrier_refs
        body.update((mutation or Mutation()).payload())
        return _build(Agent, self._request("POST", "/v1/agents", body))

    def list_agents(self) -> list[Agent]:
        return [_build(Agent, item) for item in self._items("/v1/agents")]

    def transition_agent(self, agent_ref: str, state: str, mutation: Optional[Mutation] = None) -> Agent:
        body = {"state": state}
        body.update((mutation or Mutation()).payload())
        return _build(Agent, self._request("POST", f"/v1/agents/{agent_ref}/lifecycle", body))

    def contain_agent(self, agent_ref: str, mode: str, mutation: Optional[Mutation] = None) -> ContainmentResult:
        body = {"mode": mode}
        body.update((mutation or Mutation()).payload())
        return _build(ContainmentResult, self._request("POST", f"/v1/agents/{agent_ref}/containment", body))

    # -- Identities ------------------------------------------------------

    def allocate_identity(self, namespace: str, agent_ref: str, mutation: Optional[Mutation] = None) -> AgentIdentity:
        body = {"namespace": namespace, "agentRef": agent_ref}
        body.update((mutation or Mutation()).payload())
        return _build(AgentIdentity, self._request("POST", "/v1/agent-identities", body))

    def list_identities(self, namespace: str) -> list[AgentIdentity]:
        query = urllib.parse.urlencode({"namespace": namespace})
        return [_build(AgentIdentity, item) for item in self._items(f"/v1/agent-identities?{query}")]

    def bind_identity(self, namespace: str, agent_id: str, root_ref: str, root_type: str, mutation: Optional[Mutation] = None) -> AgentIdentity:
        body = {"namespace": namespace, "authorityRootRef": root_ref, "authorityRootType": root_type}
        body.update((mutation or Mutation()).payload())
        return _build(AgentIdentity, self._request("POST", f"/v1/agent-identities/{agent_id}/binding", body))

    def transition_identity(self, namespace: str, agent_id: str, state: str, mutation: Optional[Mutation] = None) -> AgentIdentity:
        body = {"namespace": namespace, "state": state}
        body.update((mutation or Mutation()).payload())
        return _build(AgentIdentity, self._request("POST", f"/v1/agent-identities/{agent_id}/lifecycle", body))

    # -- Workloads -------------------------------------------------------

    def create_workload_registration(self, namespace: str, platform: str, selector: dict[str, str], trust_domain: str, proof_methods: list[str], mutation: Optional[Mutation] = None) -> WorkloadRegistration:
        body = {
            "namespace": namespace,
            "platform": platform,
            "selector": selector,
            "trustDomain": trust_domain,
            "allowedProofMethods": proof_methods,
        }
        body.update((mutation or Mutation()).payload())
        return _build(WorkloadRegistration, self._request("POST", "/v1/workload-registrations", body))

    def list_workload_registrations(self, namespace: str) -> list[WorkloadRegistration]:
        query = urllib.parse.urlencode({"namespace": namespace})
        return [_build(WorkloadRegistration, item) for item in self._items(f"/v1/workload-registrations?{query}")]

    def transition_workload_registration(self, workload_registration_id: str, state: str, mutation: Optional[Mutation] = None) -> WorkloadRegistration:
        body = {"state": state}
        body.update((mutation or Mutation()).payload())
        return _build(WorkloadRegistration, self._request("POST", f"/v1/workload-registrations/{workload_registration_id}/lifecycle", body))

    # -- Instances -------------------------------------------------------

    def commit_instance(self, agent_id: str, commit: InstanceCommit, mutation: Optional[Mutation] = None) -> AgentInstance:
        body = commit.payload()
        body.update((mutation or Mutation()).payload())
        return _build(AgentInstance, self._request("POST", f"/v1/agent-identities/{agent_id}/instances", body))

    def list_instances(self, namespace: str, agent_id: str) -> list[AgentInstance]:
        query = urllib.parse.urlencode({"namespace": namespace})
        return [_build(AgentInstance, item) for item in self._items(f"/v1/agent-identities/{agent_id}/instances?{query}")]

    def transition_instance(self, instance_id: str, state: str, mutation: Optional[Mutation] = None) -> AgentInstance:
        body = {"state": state}
        body.update((mutation or Mutation()).payload())
        return _build(AgentInstance, self._request("POST", f"/v1/instances/{instance_id}/lifecycle", body))

    # -- Evidence --------------------------------------------------------

    def list_evidence(self) -> list[LifecycleEvent]:
        return [_build(LifecycleEvent, item) for item in self._items("/v1/evidence")]

    def list_events(self) -> list[OutboxEvent]:
        return [_build(OutboxEvent, item) for item in self._items("/v1/events")]

    # -- Transport -------------------------------------------------------

    def _items(self, path: str) -> list[dict[str, Any]]:
        payload = self._request("GET", path)
        return list(payload.get("items") or [])

    def _request(self, method: str, path: str, body: Optional[dict[str, Any]] = None) -> dict[str, Any]:
        data = None
        headers: dict[str, str] = {}
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        if method != "GET":
            headers["Idempotency-Key"] = secrets.token_hex(16)
            headers["X-Actor"] = self.actor

        request = urllib.request.Request(self.base_url + path, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(request, timeout=self.timeout) as response:
                payload = response.read()
        except urllib.error.HTTPError as error:
            raise _api_error(error) from None

        return json.loads(payload) if payload else {}


def _api_error(error: urllib.error.HTTPError) -> APIError:
    try:
        payload = error.read()
    finally:
        error.close()

    code = message = correlation = ""
    try:
        detail = (json.loads(payload) or {}).get("error") or {}
        code = detail.get("code", "")
        message = detail.get("message", "")
        correlation = detail.get("correlationId", "")
    except Exception:  # noqa: BLE001 - fall back to the HTTP reason
        message = str(error.reason)
    return APIError(error.code, code, message, correlation)
