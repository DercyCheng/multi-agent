# Phase II — 二期规划 (Multi-Agent 平台)

本文件是针对当前仓库（`Multi-agent`）的二期（Phase II）实施规划草案。规划基于仓库内的现有文档与配置（`README.md`、`Multi-agent.md`、`config/multi-agent.yaml`），并直接映射到源码组件：`go/`（编排与网关）、`rust/agent-core/`（执行层）、`python/llm-service/`（智能层）、以及`config/`、`policies/`、`proto/` 等支持模块。

目标是把项目从目前的“平台骨架与部分实现”推进到“企业可用的首个最小可交付版本（MVP）”，聚焦稳定性、成本控制、多租户安全与可观测性。

## 高层目标（4-6 个月）

- 交付企业级可用的 Multi-Agent MVP：支持 3-5 个并行智能体、完整会话/租户隔离、令牌预算控制、基本工具（MCP）集成。
- 完成关键路径的端到端测试与 CI：单元/集成测试、工作流重放测试、故障注入测试。
- 成熟的部署与运维路径：生产级 `docker-compose`/Kubernetes 配置、监控告警、备份与恢复策略。
- 安全与合规：OPA 策略覆盖核心授权用例、WASI 沙箱强化、密钥与机密管理实践。

## 范围与不在范围

包含：

- Orchestrator（Go）: DAG 引擎完成、Temporal 工作流集成稳定、预算与并发控制、gRPC/HTTP API 完整。
- Agent Core（Rust）: WASI 沙箱、基本信念状态（BeliefState）、执行/计量、资源限制。
- LLM Service（Python）: 多供应商路由、提示缓存、模型分层（small/medium/large）与成本控制。
- 存储与索引：Postgres 模式完成、Qdrant 集合配置与测试、Redis 会话/缓存策略。
- 基础设施：生产 `docker-compose.prod.yml` 校验、Prometheus/Grafana/Jaeger 仪表盘、日志旋转与持久化。

不包含（或延后到 Phase III）：

- 区块链/钱包（Solana）深度集成（目前为可选 feature）。
- 移动 SDK、区块链记账与高级分析（复杂审计/长期仓库）。

## 里程碑与时间线（建议迭代周期：2 周/ Sprint）

- Sprint 0 (准备, 1 周)
  - 确认开发 & CI 环境；更新 `Makefile` 与 `docker-compose` 以支持本地一键启动所有服务（Postgres/Redis/Qdrant/Temporal）。
  - 定义代码所有权与分支策略（`main`/`develop`/feature）。

- Sprint 1-2 (基础稳定性, 2x2 周)
  - Orchestrator：实现并验证 DAG 执行路径、Temporal 重试/重放与健康检查（`go/orchestrator/`）。
  - 数据模型：完成并运行 DB migration（`migrations/postgres/`），并写入基础集成测试。
  - 单元测试覆盖 Orchestrator 核心模块。

- Sprint 3-4 (执行层与安全, 2x2 周)
  - Agent Core：完成 WASI 运行时资源限制、沙箱文件系统/网络隔离、执行计量（`rust/agent-core/`）。
  - OPA：完善 `policies/multiagent.rego`，覆盖会话访问控制、工具调用权限。
  - 端到端安全测试（沙箱绕过/策略违规场景）。

- Sprint 5-6 (智能层与成本控制, 2x2 周)
  - LLM Service：实现 provider manager、prompt cache、成本限制（`python/llm-service/`）。
  - 模型路由压力测试和成本模拟（注：使用 mock provider）。
  - 实装告警：当日预算 > alert_threshold 时触发告警（Prometheus + Alertmanager）。

- Sprint 7-8 (可观测性与稳定化, 2x2 周)
  - 完成 Prometheus/Grafana/Jaeger 仪表板与文档。添加关键 SLO 指标。
  - 编写故障注入测试、运行高负载 soak 测试并修复阻塞问题。

- Release (1 周)
  - 打包 Docker 镜像、生成 `docker-compose.prod.yml` 验证脚本、发布版本 & 变更日志。

总计：约 16 周（4 个月）为完整实施；若并行团队增加，可压缩为 3 个月。

## 交付物（每个里程碑结束时）

- 工作的 API 文档与示例（更新 `docs/` 中的相关文件）。
- 自动化测试（单元 + 集成 + e2e）覆盖率报告。 
- 可运行的 `docker-compose` 生产配置与部署指南。
- OPA 策略覆盖文档与示例用例。

## 风险与缓解措施

- 供应商 API 费用不可控：使用 mock provider 进行压力测试，并实现 token-budget 限制与每日预算告警。
- WASI 沙箱突破或资源泄漏：限制内存/CPU、引入定期内存回收与监控（agent_core.memory_pool），并设置 circuit-breakers。
- Temporal 工作流错误或回放不一致：增加 deterministic tests、回放日志记录与更严格的版本控制。
- 多租户数据隔离错误：编写集成测试覆盖租户边界、严格 DB schema 限制与 OPA 策略。

## 资源估算（示例）

- 小型团队（3-4 人）：4 个月
  - 1x 后端 Go 开发（Orchestrator / API Gateway）
  - 1x Rust 开发（Agent Core / WASI）
  - 1x Python 开发（LLM 服务 / Provider 集成）
  - 0.5x DevOps（CI/CD、监控、发布）

- 中型团队（6 人）：2-3 个月（并行化）

## 验收标准（可量化）

1. 功能：能够在单个部署上完成端到端任务，从 API 提交到 Agent 执行并返回结果。例：`/api/v1/tasks` -> Orchestrator -> Agent Core -> LLM -> 完成。
2. 性能：在默认配置与 5 个并行智能体下，冷启动响应 < 2s，99p 请求延迟 < 5s（受限于 provider latency）。
3. 成本控制：token 使用计量准确，预算超限触发保护并拒绝超额请求。
4. 安全：所有代码执行均在 WASI 沙箱或受控 Python provider 中运行；关键 OPA 策略通过回归测试。
5. 可观测性：Prometheus 指标齐全（任务成功率、智能体延迟、token 消耗），并有至少 3 个 Grafana 仪表板与告警规则。

## 映射到现有仓库组件（优先级）

- 高优先级（必须）：
  - `go/orchestrator/` — DAG、Temporal、budget 管理 (主要实现责任）
  - `rust/agent-core/` — WASI、执行、资源限制（安全关键）
  - `python/llm-service/` — provider 管理、prompt cache、成本控制
  - `migrations/postgres/` — 数据模型完成与迁移脚本
  - `config/multi-agent.yaml` — 环境与 feature flags 定义
  - `policies/multiagent.rego` — OPA 策略（授权/工具调用控制）

- 中等优先级：
  - `proto/` — gRPC 协议定义的最终稳定化
  - `go/api-gateway/` — 认证、速率限制、路由
  - `python/.../integrations/` — Provider adapters
  - `docs/` — 使用说明、运维文档

- 低优先级（Phase III）：
  - `security-service/` 深度 RBAC 与审计扩展
  - 区块链/钱包集成（`rust/agent-core` 扩展）

## 具体任务清单（建议拆分为 GitHub Issues）

每项任务建议包含：标题、描述、验收标准、估时（人-天）、优先级、负责人。下面是首批候选 Issue 列表：

1. Orchestrator: 完成 Workflow DAG 重放与回放测试
   - 验收：能在回放模式下重复执行已完成 workflow，并输出一致性报告。
   - 估时：5 天

2. Agent Core: 强化 WASI 限制（memory/cpu/timeout）并添加 sandbox 逃逸单元测试
   - 验收：注入恶意/无限循环代码不会导致宿主崩溃，执行被强制中止。
   - 估时：6 天

3. LLM Service: 实现 provider manager + prompt cache（mock providers）
   - 验收：在启用缓存时重复请求命中率 >= 60%（模拟场景），provider 切换正常。
   - 估时：5 天

4. CI: 添加 pipeline，包含单元测试、lint、build、basic e2e（docker compose up）
   - 验收：PR 提交后 CI 运行并通过（或失败并给出错误输出）。
   - 估时：4 天

5. Observability: 添加 Prometheus metrics 到 Orchestrator & Agent Core，并提供基础 Grafana dashboard
   - 验收：关键指标可视化，dashboard 模版 checked in。
   - 估时：3 天

6. Security/OPA: 补全 `policies/multiagent.rego` 并添加策略测试用例
   - 验收：策略测试覆盖常见访问场景并通过。
   - 估时：3 天

7. Release: 产出 `docker-compose.prod.yml` 校验脚本与 deploy playbook
   - 验收：通过一键部署脚本启动所有关键服务在本地或干净服务器。
   - 估时：4 天

## 快速开始（开发者任务优先级）

1. 本地环境：确保 `.env` 中配置了所有 provider mock keys，并且 `docker compose up -d` 能够启动 postgres/redis/qdrant/temporal。
2. 运行 DB migration：`migrations/postgres/` 下的 SQL 文件执行一次。
3. 启动 Orchestrator（开发模式）并运行一个示例任务，观察 Prometheus 指标。

## 下步（由我执行或协助的事项）

- 我可以根据上面的 Issue 列表为仓库生成具体的 `ISSUE` 文本，或直接在你的 issue tracker 中批量创建（需要访问权限）。
- 我可以继续把第 3 项 todo 完成：把 action items 转成 `docs/` 下的任务板、为每个任务生成 PR 模板和 CI job 模板。

---

作者：规划自动生成器（基于仓库上下文）
最后更新：2025-09-20

## Generated Issues (first batch)

下面是可以直接复制到 GitHub Issue 的模板（第一批）。每个 Issue 包含标题、描述、验收标准和估时建议。

### Issue: Orchestrator - Workflow DAG replay and deterministic tests
Description:
- Implement and test workflow replay for Orchestrator. Ensure deterministic replay of completed workflows and provide a consistency report.

Acceptance Criteria:
- A replay mode exists that can re-run an existing workflow and compare outputs.
- Unit/integration tests demonstrate replay determinism for 3 representative workflows.
- Documentation updated in `docs/`.

Estimate: 5 days

---

### Issue: Agent Core - Harden WASI sandbox (memory/cpu/timeout)
Description:
- Enforce memory, CPU and timeout limits in the WASI sandbox. Add tests that attempt resource exhaustion or infinite loops and verify the host remains stable.

Acceptance Criteria:
- Sandbox kills runaway code within configured limits.
- Tests for sandbox escape and resource exhaustion are added and passing.

Estimate: 6 days

---

### Issue: LLM Service - Provider manager and prompt cache (with mock providers)
Description:
- Implement provider manager and prompt cache. Wire mock providers to simulate latency/cost and validate routing and cache hit rates.

Acceptance Criteria:
- Provider manager supports selection by tier and priority.
- Prompt cache stores and retrieves results; simulated cache hit >= 60% in test scenarios.

Estimate: 5 days

---

### Issue: CI - Add pipeline for build/test/lint/e2e
Description:
- Create CI pipeline that runs unit tests, linting, build for Go/Rust/Python and a basic e2e using `docker compose up`.

Acceptance Criteria:
- PR trigger runs the pipeline; failures block merge.

Estimate: 4 days

---

### Issue: Observability - Add Prometheus metrics and Grafana dashboards
Description:
- Instrument Orchestrator and Agent Core for key metrics and provide Grafana dashboards as JSON templates.

Acceptance Criteria:
- Metrics exposed and scraped by Prometheus in local setup.
- Grafana dashboards committed under `docs/monitoring/`.

Estimate: 3 days

---

### Issue: Security/OPA - Complete `policies/multiagent.rego` and add tests
Description:
- Extend OPA policies to cover session/tenant access and tool invocation permissions. Add unit tests for policies.

Acceptance Criteria:
- Coverage for common access scenarios and CI-run policy tests.

Estimate: 3 days

---

### Issue: Release - Produce `docker-compose.prod.yml` validation and deploy playbook
Description:
- Provide scripts to validate `docker-compose.prod.yml`, build images, and a playbook for one-click deploy to a clean server.

Acceptance Criteria:
- One script builds images and deploys to a clean VM; health checks pass.

Estimate: 4 days

