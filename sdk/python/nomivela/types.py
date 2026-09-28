"""Response and request types for the NOMIVELA Agent Registry API."""

from __future__ import annotations

from dataclasses import dataclass, field, fields
from datetime import datetime
from typing import Any, Optional


def _json(name: str):
    return field(default=None, metadata={"json": name})


def _build(cls, data: dict[str, Any]):
    kwargs: dict[str, Any] = {}
    for item in fields(cls):
        key = item.metadata.get("json", item.name)
        if key in data and data[key] is not None:
            kwargs[item.name] = data[key]
    return cls(**kwargs)


@dataclass
class Namespace:
    namespace: Optional[str] = _json("namespace")
    parent: Optional[str] = _json("parent")
    authority_root_ref: Optional[str] = _json("authorityRootRef")
    status: Optional[str] = _json("status")
    namespace_epoch: Optional[int] = _json("namespaceEpoch")


@dataclass
class Agent:
    agent_ref: Optional[str] = _json("agentRef")
    name: Optional[str] = _json("name")
    purpose: Optional[str] = _json("purpose")
    sponsor_ref: Optional[str] = _json("sponsorRef")
    owner_ref: Optional[str] = _json("ownerRef")
    risk_class: Optional[str] = _json("riskClass")
    agent_class: Optional[str] = _json("agentClass")
    carrier_refs: Optional[list[str]] = _json("carrierRefs")
    state: Optional[str] = _json("state")
    agent_epoch: Optional[int] = _json("agentEpoch")


@dataclass
class AgentIdentity:
    namespace: Optional[str] = _json("namespace")
    agent_id: Optional[str] = _json("agentId")
    agent_ref: Optional[str] = _json("agentRef")
    state: Optional[str] = _json("state")
    identity_epoch: Optional[int] = _json("identityEpoch")
    authority_root_ref: Optional[str] = _json("authorityRootRef")
    authority_root_type: Optional[str] = _json("authorityRootType")


@dataclass
class WorkloadRegistration:
    workload_registration_id: Optional[str] = _json("workloadRegistrationId")
    namespace: Optional[str] = _json("namespace")
    platform: Optional[str] = _json("platform")
    selector: Optional[dict[str, str]] = _json("selector")
    trust_domain: Optional[str] = _json("trustDomain")
    allowed_proof_methods: Optional[list[str]] = _json("allowedProofMethods")
    proof_requirements: Optional[dict[str, Any]] = _json("proofRequirements")
    status: Optional[str] = _json("status")
    workload_epoch: Optional[int] = _json("workloadEpoch")


@dataclass
class AgentInstance:
    instance_id: Optional[str] = _json("instanceId")
    namespace: Optional[str] = _json("namespace")
    agent_id: Optional[str] = _json("agentId")
    workload_registration_id: Optional[str] = _json("workloadRegistrationId")
    workload_id: Optional[str] = _json("workloadId")
    artifact_digest: Optional[str] = _json("artifactDigest")
    attestation_ref: Optional[str] = _json("attestationRef")
    lease_expires_at: Optional[str] = _json("leaseExpiresAt")
    state: Optional[str] = _json("state")
    generation: Optional[int] = _json("generation")


@dataclass
class LifecycleEvent:
    event_id: Optional[str] = _json("eventId")
    object_type: Optional[str] = _json("objectType")
    object_id: Optional[str] = _json("objectId")
    previous_state: Optional[str] = _json("previousState")
    new_state: Optional[str] = _json("newState")
    epoch_kind: Optional[str] = _json("epochKind")
    epoch: Optional[int] = _json("epoch")
    actor: Optional[str] = _json("actor")
    reason: Optional[str] = _json("reason")
    evidence_ref: Optional[str] = _json("evidenceRef")
    occurred_at: Optional[str] = _json("occurredAt")


@dataclass
class OutboxEvent:
    event_id: Optional[str] = _json("eventId")
    event_type: Optional[str] = _json("eventType")
    aggregate_type: Optional[str] = _json("aggregateType")
    aggregate_id: Optional[str] = _json("aggregateId")
    sequence: Optional[int] = _json("sequence")
    cursor: Optional[int] = _json("cursor")
    payload_version: Optional[int] = _json("payloadVersion")
    payload: Optional[dict[str, Any]] = _json("payload")
    occurred_at: Optional[str] = _json("occurredAt")
    attempts: Optional[int] = _json("attempts")


@dataclass
class EventPage:
    items: Optional[list[OutboxEvent]] = _json("items")
    next_cursor: Optional[int] = _json("nextCursor")


@dataclass
class RegistryContext:
    """A consistent point-in-time Registry read for one Agent identity."""

    namespace: Optional[Any] = _json("namespace")
    agent: Optional[Any] = _json("agent")
    identity: Optional[Any] = _json("identity")
    workload_registration: Optional[Any] = _json("workloadRegistration")
    instance: Optional[Any] = _json("instance")


@dataclass
class DiscoveryDocument:
    namespace: Optional[str] = _json("namespace")
    registry_endpoint: Optional[str] = _json("registryEndpoint")
    issuer: Optional[str] = _json("issuer")
    jwks_uri: Optional[str] = _json("jwksUri")
    supported_proof_profiles: Optional[list[str]] = _json("supportedProofProfiles")
    supported_artifact_types: Optional[list[str]] = _json("supportedArtifactTypes")
    key_rotation: Optional[str] = _json("keyRotation")
    discovery_version: Optional[str] = _json("discoveryVersion")
    issued_at: Optional[str] = _json("issuedAt")
    expires_at: Optional[str] = _json("expiresAt")
    signing_kid: Optional[str] = _json("signingKid")
    alg: Optional[str] = _json("alg")
    signature: Optional[str] = _json("signature")


@dataclass
class JWKS:
    keys: Optional[list[dict[str, Any]]] = _json("keys")


@dataclass
class ContainmentResult:
    agent_ref: Optional[str] = _json("agentRef")
    mode: Optional[str] = _json("mode")
    agent_state: Optional[str] = _json("agentState")
    agent_epoch: Optional[int] = _json("agentEpoch")
    identities: Optional[list[str]] = _json("identities")
    instances: Optional[list[str]] = _json("instances")


@dataclass
class InstanceCommit:
    namespace: str
    workload_registration_id: str
    workload_id: str
    artifact_digest: str
    attestation_ref: str
    lease_expires_at: datetime | str

    def payload(self) -> dict[str, Any]:
        lease = self.lease_expires_at
        if isinstance(lease, datetime):
            lease = lease.isoformat().replace("+00:00", "Z")
        return {
            "namespace": self.namespace,
            "workloadRegistrationId": self.workload_registration_id,
            "workloadId": self.workload_id,
            "artifactDigest": self.artifact_digest,
            "attestationRef": self.attestation_ref,
            "leaseExpiresAt": lease,
        }
