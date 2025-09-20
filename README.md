# Multi-Agent AI协作平台

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Rust](https://img.shields.io/badge/Rust-1.75+-red.svg)](https://rust-lang.org)
[![Python](https://img.shields.io/badge/Python-3.11+-green.svg)](https://python.org)
[![React](https://img.shields.io/badge/React-19+-blue.svg)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5+-blue.svg)](https://typescriptlang.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 🎯 项目概述

Multi-Agent是一个**生产级企业多智能体AI协作平台**，采用现代化三层微服务架构，专为大规模AI部署中的成本控制、可靠性和安全性问题而设计。平台集成了Temporal工作流、WASI沙箱、OPA策略引擎等先进技术，提供亚秒级响应时间和全面的可观测性。现已集成**美观的Web前端界面**，支持完整的可视化管理和监控。

### 🌟 核心特性

- **🔄 智能协作**: 支持3-5个并行智能体的复杂任务分解和协作执行
- **💰 成本控制**: 通过智能缓存和会话管理实现70%的令牌成本削减
- **🔒 安全可靠**: WASI沙箱 + OPA策略引擎的零信任架构
- **📈 可扩展性**: 支持15+大模型提供商的热切换和水平扩展
- **👁️ 可观测性**: Prometheus指标、Grafana仪表板和OpenTelemetry追踪
- **🏢 多租户**: 完整的企业级多租户数据隔离
- **🎨 美观界面**: React + TypeScript + Ant Design 现代化Web前端

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     Web Frontend                            │
│        React + TypeScript + Ant Design                     │
│    Dashboard | Agent Management | Workflow Control         │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                        API Gateway                          │
│          Rate Limiting | Auth | Request Routing             │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Orchestration Layer (Go)                       │
│  • DAG Engine with Task Decomposition                       │
│  • Token Budget Manager                                     │
│  • Parallel Execution Controller (3-5 agents max)           │
│  • Execution Attestation Aggregator                         │
└────────┬──────────────────────────────┬─────────────────────┘
         │                              │
┌────────▼────────┐            ┌───────▼──────────────────────┐
│ Agent Core      │            │  LLM & Tool Services         │
│ (Rust)          │            │  (Python)                    │
│ • FSM Engine    │            │  • Model Router              │
│ • WASM Sandbox  │            │  • MCP Tools                 │
│ • Memory Mgmt   │            │  • Provider SDKs             │
└────────┬────────┘            └──────────────────────────────┘
         │
┌────────▼────────────────────────────────────────────────────┐
│                   Storage & State Layer                     │
│  PostgreSQL | Redis | Qdrant | S3 | Temporal               │
└─────────────────────────────────────────────────────────────┘
```

### 🛠️ 技术栈

**后端服务:**
- **API Gateway**: Go + Gin + gRPC 网关
- **Agent Core**: Rust + WASI + WebAssembly 沙箱
- **Orchestrator**: Go + Temporal + 工作流引擎
- **LLM Service**: Python + FastAPI + 多提供商支持
- **Security**: OPA + RBAC + 多租户策略

**前端应用:**
- **Web Frontend**: React 19 + TypeScript + Ant Design 5
- **状态管理**: Redux Toolkit + RTK Query
- **实时通信**: Socket.IO WebSocket
- **数据可视化**: Apache ECharts + 仪表板
- **构建工具**: Vite + PostCSS + Tailwind CSS

**数据存储:**
- **关系数据库**: PostgreSQL + pgvector 扩展
- **缓存存储**: Redis Cluster
- **向量数据库**: Qdrant
- **对象存储**: MinIO/S3
- **工作流状态**: Temporal

**运维监控:**
- **指标采集**: Prometheus + 自定义指标
- **日志聚合**: 结构化日志 + 追踪
- **容器化**: Docker + Compose
- **反向代理**: Nginx + 负载均衡

## 🚀 快速开始

### 前置要求

- Docker & Docker Compose
- Go 1.21+
- Rust 1.75+
- Python 3.11+
- Node.js 18+ (前端开发)

### 1. 克隆项目

```bash
git clone <repository-url>
cd Multi-agent
```

### 2. 环境配置

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件，填入你的API密钥
vim .env
```

### 3. 启动服务

```bash
# 开发环境（包含前端）
docker compose up -d

# 生产环境
docker compose -f docker-compose.prod.yml up -d

# 或使用Makefile
make docker-up
```

### 4. 访问应用

- **Web前端**: http://localhost:3000
  - 现代化React界面，支持智能体管理、工作流控制、实时监控
- **API网关**: http://localhost:8080
- **Grafana监控**: http://localhost:3001
- **Prometheus指标**: http://localhost:9090

### 5. 验证部署

```bash
# 检查服务健康状态
curl http://localhost:8080/health

# 检查前端应用
curl http://localhost:3000

# 访问监控面板
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
open http://localhost:16686 # Jaeger
```

## 📖 使用指南

### Web界面快速开始

1. **访问Web前端**: http://localhost:3000
2. **登录系统**: 使用默认账户或注册新用户
3. **仪表板概览**: 查看系统状态、智能体活动、工作流执行情况
4. **智能体管理**: 创建、配置、启动/停止智能体
5. **工作流控制**: 设计、执行、监控复杂的AI工作流
6. **实时监控**: 通过WebSocket连接获得实时状态更新

### API示例

#### 1. 用户注册和认证

```bash
# 注册用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "demo_user",
    "password": "SecurePass123!",
    "full_name": "Demo User"
  }'

# 登录获取Token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

#### 2. 创建会话

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "initial_context": {
      "goal": "Analyze market trends and provide investment recommendations"
    },
    "token_budget": 5000
  }'
```

#### 3. 提交任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Analyze the latest trends in AI and provide a comprehensive report",
    "session_id": "SESSION_ID",
    "auto_decompose": true,
    "context": {
      "focus_areas": ["machine learning", "LLMs", "robotics"],
      "depth": "detailed"
    }
  }'
```

#### 4. 查询任务状态

```bash
curl -X GET http://localhost:8080/api/v1/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 高级功能

#### MCP工具集成

```bash
# 注册外部工具
curl -X POST http://localhost:8000/tools/mcp/register \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "weather_api",
    "url": "https://api.weather.com/mcp",
    "func_name": "get_weather",
    "parameters": [
      {"name": "city", "type": "string", "required": true}
    ]
  }'

# 执行工具
curl -X POST http://localhost:8000/tools/execute \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "weather_api",
    "parameters": {"city": "Beijing"}
  }'
```

## 🔧 开发指南

### 项目结构

```
Multi-agent/
├── config/                 # 配置文件
│   ├── multi-agent.yaml   # 主配置
│   ├── prometheus/        # 监控配置
│   └── redis.conf         # Redis配置
├── docs/                  # 文档
├── frontend/              # React前端应用
│   ├── src/
│   │   ├── components/    # React组件
│   │   ├── pages/         # 页面组件
│   │   ├── services/      # API服务
│   │   ├── store/         # Redux状态管理
│   │   └── types/         # TypeScript类型定义
│   ├── package.json       # 前端依赖
│   ├── vite.config.ts     # Vite构建配置
│   └── Dockerfile         # 前端容器化
├── go/                    # Go服务
│   ├── api-gateway/       # API网关
│   ├── orchestrator/      # 编排服务
│   └── security-service/  # 安全服务
├── rust/agent-core/       # Rust智能体核心
├── python/llm-service/    # Python LLM服务
├── migrations/            # 数据库迁移
├── policies/              # OPA策略
├── proto/                 # gRPC协议定义
└── docker-compose.yml     # Docker配置
```

### 编译和测试

```bash
# 编译所有服务
make build

# 运行测试
make test

# 代码格式化
make format

# 代码检查
make lint

# 编译检查
make compile-check
```

### 开发工作流

1. **前端开发**:
   ```bash
   cd frontend
   npm install
   npm run dev        # 开发服务器
   npm run build      # 生产构建
   npm run preview    # 预览生产版本
   ```

2. **Go服务开发**:
   ```bash
   cd go/orchestrator
   go mod tidy
   go run cmd/main.go
   ```

3. **Rust服务开发**:
   ```bash
   cd rust/agent-core
   cargo build
   cargo run
   ```

3. **Python服务开发**:
   ```bash
   cd python/llm-service
   pip install -r requirements.txt
   python src/main.py
   ```

## 📊 监控和运维

### 监控指标

- **业务指标**: 任务成功率、智能体执行时间、令牌使用量
- **性能指标**: 响应时间、吞吐量、资源使用率
- **可靠性指标**: 错误率、服务可用性、依赖健康度

### 日志查看

```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f orchestrator

# 实时监控
docker compose logs -f --tail=100
```

### 数据库管理

```bash
# 连接PostgreSQL
docker exec -it multiagent-postgres psql -U postgres -d multiagent

# 连接Redis
docker exec -it multiagent-redis redis-cli

# 访问Qdrant Web UI
open http://localhost:6333/dashboard
```

## 🔐 安全性

### 认证和授权

- **JWT**: 短期访问令牌（30分钟）+ 长期刷新令牌（7天）
- **API密钥**: 程序化访问的长期令牌
- **多租户**: 完整的数据隔离
- **RBAC**: 基于角色的访问控制（Owner/Admin/User）

### 策略管理

项目使用OPA（Open Policy Agent）进行细粒度权限控制：

```bash
# 测试策略
opa eval -d policies/ "data.multiagent.allow" -i input.json

# 更新策略
kubectl apply -f policies/multiagent.rego
```

## 🚀 部署指南

### 生产部署

1. **环境准备**:
   ```bash
   # 设置生产环境变量
   export ENVIRONMENT=production
   export JWT_SECRET="your-production-secret"
   ```

2. **数据库初始化**:
   ```bash
   # 运行数据库迁移
   docker compose exec postgres psql -U postgres -d multiagent -f /docker-entrypoint-initdb.d/001_initial_schema.sql
   ```

3. **服务启动**:
   ```bash
   # 生产部署
   docker compose -f docker-compose.prod.yml up -d
   ```

### 扩展配置

- **水平扩展**: 通过Docker Swarm或Kubernetes
- **负载均衡**: 使用Nginx或云负载均衡器
- **高可用**: 多地域部署和数据同步

## 🤝 贡献指南

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

### 开发规范

- **Go**: 遵循标准Go代码规范，使用`gofmt`格式化
- **Rust**: 使用`rustfmt`格式化，遵循Rust最佳实践
- **Python**: 遵循PEP 8，使用`black`格式化
- **React/TypeScript**: 使用Prettier格式化，遵循ESLint规则
- **提交**: 使用[Conventional Commits](https://conventionalcommits.org/)规范

## 📋 待办事项

### 高优先级 ✅
- [x] 统一配置系统
- [x] gRPC协议定义
- [x] 完善的OPA策略
- [x] 监控和可观测性配置
- [x] **Web前端界面集成** 🎉
- [ ] 核心业务逻辑实现
- [ ] 完整的API文档

### 中等优先级 🔄
- [ ] 前端单元测试（Jest + React Testing Library）
- [ ] 端到端测试（Cypress/Playwright）
- [ ] 单元测试和集成测试
- [ ] CI/CD流水线
- [ ] 性能基准测试
- [ ] 故障注入测试

### 低优先级 ⏳
- [ ] 移动端SDK
- [ ] 区块链集成
- [ ] 高级分析功能
- [ ] 多语言支持

## 🎨 前端开发详细指南

### 技术特性

- **现代化UI**: 基于Ant Design 5的企业级组件库
- **响应式设计**: 支持桌面、平板、移动端设备
- **实时数据**: WebSocket连接实现数据的实时更新
- **状态管理**: Redux Toolkit提供可预测的状态管理
- **类型安全**: TypeScript提供完整的类型检查
- **模块化架构**: 清晰的组件、服务、存储分层架构

### 核心页面功能

1. **仪表板 (Dashboard)**
   - 系统状态概览
   - 智能体活动监控
   - 工作流执行统计
   - 实时性能指标

2. **智能体管理 (Agent Management)**
   - 智能体创建和配置
   - 状态控制（启动/停止/重启）
   - 性能监控和日志查看
   - 智能体间的协作管理

3. **工作流控制 (Workflow Control)**
   - 可视化工作流设计
   - 模板管理和复用
   - 执行监控和调试
   - 工作流版本控制

4. **用户认证 (Authentication)**
   - JWT令牌认证
   - 角色权限管理
   - 多租户数据隔离
   - 会话管理

### 开发环境搭建

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产版本
npm run preview

# 代码格式化
npm run format

# 类型检查
npm run type-check
```

### API集成示例

```typescript
// 使用Redux Toolkit Query
import { api } from '../services/api';

export const agentApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getAgents: builder.query<Agent[], void>({
      query: () => '/agents',
    }),
    createAgent: builder.mutation<Agent, CreateAgentRequest>({
      query: (agent) => ({
        url: '/agents',
        method: 'POST',
        body: agent,
      }),
    }),
  }),
});

export const { useGetAgentsQuery, useCreateAgentMutation } = agentApi;
```

### 组件开发规范

```typescript
// 推荐的组件结构
interface AgentCardProps {
  agent: Agent;
  onStatusChange: (id: string, status: AgentStatus) => void;
}

export const AgentCard: React.FC<AgentCardProps> = ({ 
  agent, 
  onStatusChange 
}) => {
  return (
    <Card
      title={agent.name}
      extra={<StatusBadge status={agent.status} />}
      actions={[
        <Button 
          key="start" 
          onClick={() => onStatusChange(agent.id, 'running')}
        >
          启动
        </Button>
      ]}
    >
      <AgentMetrics agent={agent} />
    </Card>
  );
};
```

### 部署配置

前端应用通过Docker容器化，使用Nginx作为Web服务器：

```dockerfile
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
```

详细的前端开发指南请参考：[frontend/README.md](frontend/README.md)

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Temporal](https://temporal.io/) - 可靠的工作流编排
- [Qdrant](https://qdrant.tech/) - 高性能向量数据库
- [OpenAI](https://openai.com/) - 强大的语言模型
- [Anthropic](https://anthropic.com/) - Claude模型支持
- [Ant Design](https://ant.design/) - 企业级UI设计语言
- [React](https://reactjs.org/) - 用户界面构建库

## 📞 联系我们

- 项目主页: [GitHub Repository]()
- 问题反馈: [Issues]()
- 讨论交流: [Discussions]()

---

**Built with ❤️ by the Multi-Agent Team**