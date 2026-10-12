"""NOMIVELA Agent Registry SDK for Python."""

from .client import APIError, Client, Mutation
from .discovery import (
    DiscoveryVerificationError,
    canonical_discovery,
    jwk_public_key,
    verify_discovery,
    verify_discovery_with_jwks,
)
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
)

__all__ = [
    "APIError",
    "Agent",
    "AgentIdentity",
    "AgentInstance",
    "Client",
    "ContainmentResult",
    "DiscoveryVerificationError",
    "InstanceCommit",
    "LifecycleEvent",
    "Mutation",
    "Namespace",
    "OutboxEvent",
    "WorkloadRegistration",
    "canonical_discovery",
    "jwk_public_key",
    "verify_discovery",
    "verify_discovery_with_jwks",
]
