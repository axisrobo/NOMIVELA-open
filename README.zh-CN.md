# NOMIVELA

[English](README.md) | [简体中文](README.zh-CN.md)

**Agent 注册中心与命名空间权威（The Agent Registry and Namespace Authority）。**

NOMIVELA 是面向组织内 AI Agent 的系统记录（system of record），拥有 Agent 记录、Agent 身份记录、权威命名空间、不可变权威绑定、工作负载注册、Agent 实例、生命周期 epoch、发现（Discovery）、证据（Evidence）与注册表事件。

## NOMIVELA 解决的问题

一个 Agent 的业务身份、安全身份、运行时部署与授权，通常被混在同一个应用或身份提供方里。由此带来三个问题：

- Agent 的生命周期无法与其所持有的凭据解耦、独立演进；
- 变更权威根（authority root）不安全，因为历史会被就地重写；
- 同一个 Agent 无法在跨云、跨组织、私有环境中被一致地发现。

## NOMIVELA 做什么

NOMIVELA 通过为每个 Agent 提供一条权威、持久、以 PostgreSQL 为后端的注册记录，把上述关注点分离：

- **命名空间与不可变权威绑定**：权威根可以变更，而无需重写历史。
- **生命周期 epoch 与原子化的证据/发件箱（outbox）事件**：每次状态变化都有序且可验证。
- **发现元数据（Discovery metadata）**：Agent 可被一致地发现。
- **严格边界**：NOMIVELA 不签发凭据或令牌、不认证工作负载、不做授权决策、不签发 Execution Grant。这些分别属于 EIDOVELA（身份与凭据）和 AEGIVELA（授权）。

## 本仓库：NOMIVELA Open

NOMIVELA Open 是 NOMIVELA 的公开发布仓库，包含：

- 公开 API 契约与 JSON Schema（`contracts/`）
- 一致性测试夹具（`conformance/`）
- Go、Python、Java SDK（`sdk/`）
- 命令行客户端（`cli/nomivela`）
- 可运行的示例（`examples/`）
- 可再分发的 NOMIVELA Core 二进制文件，见 [Releases](https://github.com/axisrobo/NOMIVELA-open/releases) 页面（`dist/`）

本仓库不包含 Core 源码、企业版（Enterprise Edition）代码、内部设计文档或任何凭据。

运行 API 见 `contracts/core-api.openapi.yaml`；Agent Registry 契约见 `contracts/agent-registry-v0.1.openapi.yaml`。

## 仓库家族

| 仓库 | 可见性 | 用途 | 许可证 |
| --- | --- | --- | --- |
| [NOMIVELA](https://github.com/axisrobo/NOMIVELA) | 源码可见 | Core 注册中心实现（Go、PostgreSQL） | AGPL-3.0-or-later |
| **NOMIVELA-open**（本仓库） | 公开 | 契约、Schema、SDK、CLI、示例、一致性夹具、Core 二进制 | Apache-2.0 |
| NOMIVELA-ee | 私有 | 企业版能力与内部设计 | Enterprise License |

## 许可证

除子目录另有说明外，内容采用 Apache-2.0 许可。见 `LICENSE`。

## 发布

版本采用 `major.minor.patch`，例如 `1.2.0`，发布标签为 `v<major.minor.patch>`。Core 与 Open 始终使用相同的版本号与标签名。
