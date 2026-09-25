# NOMIVELA

[English](README.md) | [简体中文](README.zh-CN.md)

**The Agent Registry and Namespace Authority.**

NOMIVELA is the system of record for AI agents across an organization. It owns the Agent Record, Agent Identity Record, Authority Namespace, immutable Authority Binding, Workload Registration, Agent Instance, lifecycle epochs, Discovery, Evidence, and registry events.

## The problem NOMIVELA solves

An agent's business identity, security identity, runtime deployment, and authorization are usually conflated in a single application or identity provider. As a result:

- an agent's lifecycle cannot change independently of the credentials it holds;
- changing an authority root is unsafe, because history is rewritten in place;
- the same agent cannot be discovered consistently across clouds, organizations, and private environments.

## What NOMIVELA does

NOMIVELA separates those concerns by giving every agent one authoritative, durable, PostgreSQL-backed registry entry:

- **Namespaces and immutable authority bindings**, so authority roots can change without rewriting history.
- **Lifecycle epochs and atomic evidence/outbox events**, so every state change is ordered and verifiable.
- **Discovery metadata**, so agents can be found consistently.
- **A strict boundary**: NOMIVELA does not issue credentials or tokens, authenticate workloads, make authorization decisions, or issue Execution Grants. Those belong to EIDOVELA (identity and credentials) and AEGIVELA (authorization).

## This repository: NOMIVELA Open

NOMIVELA Open is the public distribution repository for NOMIVELA. It contains:

- Public API contracts and JSON Schemas (`contracts/`)
- Conformance fixtures (`conformance/`)
- Go, Python, and Java SDKs (`sdk/`)
- A command-line client (`cli/nomivela`)
- Runnable examples (`examples/`)
- Redistributable NOMIVELA Core binaries on the [Releases](https://github.com/axisrobo/NOMIVELA-open/releases) page (`dist/`)

It does not contain Core source code, Enterprise Edition code, internal design documents, or credentials.

The operational API is described in `contracts/core-api.openapi.yaml`; the Agent Registry contract is `contracts/agent-registry-v0.1.openapi.yaml`.

## Repository family

| Repository | Visibility | Purpose | License |
| --- | --- | --- | --- |
| [NOMIVELA](https://github.com/axisrobo/NOMIVELA) | Source available | Core registry implementation (Go, PostgreSQL) | AGPL-3.0-or-later |
| **NOMIVELA-open** (this repository) | Public | Contracts, schemas, SDKs, CLI, examples, conformance fixtures, Core binaries | Apache-2.0 |
| NOMIVELA-ee | Private | Enterprise features and internal design | Enterprise License |

## License

Unless a subdirectory states otherwise, the contents are licensed under Apache-2.0. See `LICENSE`.

## Releases

Versions use `major.minor.patch`, for example `1.2.0`. Release tags are `v<major.minor.patch>`. Core and Open always share the same version number and tag name.
