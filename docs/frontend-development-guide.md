# Multi-Agent 前端开发指南

## 概述

Multi-Agent 前端是一个基于 React + TypeScript + Ant Design 构建的现代化 Web 应用程序，提供了美观且功能丰富的用户界面来管理和监控多代理系统。

## 技术栈

### 核心技术
- **React 18** - 前端框架
- **TypeScript** - 类型安全的 JavaScript
- **Vite** - 快速的构建工具
- **Ant Design 5** - 企业级 UI 组件库

### 状态管理
- **Redux Toolkit** - 状态管理
- **RTK Query** - API 状态管理

### 样式和布局
- **Tailwind CSS** - 实用性优先的 CSS 框架
- **Ant Design** - 组件样式

### 图表和可视化
- **Apache ECharts** - 数据可视化
- **echarts-for-react** - React 集成

### 实时通信
- **Socket.IO Client** - WebSocket 实时通信

### 路由
- **React Router v6** - 客户端路由

## 项目结构

```
frontend/
├── public/                 # 静态资源
├── src/
│   ├── components/        # 可复用组件
│   │   └── Layout/       # 布局组件
│   ├── pages/            # 页面组件
│   │   ├── Auth/         # 认证页面
│   │   ├── Agents/       # Agent 管理页面
│   │   └── Workflows/    # 工作流管理页面
│   ├── services/         # API 服务层
│   ├── store/            # Redux 状态管理
│   │   └── slices/       # Redux 切片
│   ├── types/            # TypeScript 类型定义
│   ├── utils/            # 工具函数
│   ├── hooks/            # 自定义 React Hooks
│   ├── App.tsx           # 主应用组件
│   └── main.tsx          # 应用入口点
├── package.json          # 依赖配置
├── vite.config.ts        # Vite 配置
├── tailwind.config.js    # Tailwind 配置
├── tsconfig.json         # TypeScript 配置
├── Dockerfile           # Docker 镜像配置
├── nginx.conf           # Nginx 配置
└── README.md            # 项目文档
```

## 开发环境设置

### 前置要求
- Node.js 18+ 
- npm 或 yarn
- Git

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd Multi-agent/frontend
   ```

2. **安装依赖**
   ```bash
   npm install
   # 或
   yarn install
   ```

3. **环境配置**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，配置 API 地址等
   ```

4. **启动开发服务器**
   ```bash
   npm run dev
   # 或
   yarn dev
   ```

   应用将在 `http://localhost:3000` 启动

### 开发工具配置

推荐使用 Visual Studio Code 并安装以下扩展：
- ES7+ React/Redux/React-Native snippets
- TypeScript Hero
- Tailwind CSS IntelliSense
- Prettier - Code formatter
- ESLint

## 核心功能

### 1. 用户认证
- 登录/登出
- JWT Token 管理
- 权限控制

### 2. 仪表板
- 系统概览
- 实时指标
- 性能监控
- 活动日志

### 3. Agent 管理
- Agent 列表查看
- 创建/编辑/删除 Agent
- Agent 状态监控
- 性能指标展示

### 4. 工作流管理
- 工作流列表查看
- 工作流创建和编辑
- 执行历史查看
- 实时执行监控

### 5. 实时通信
- WebSocket 连接管理
- 实时状态更新
- 事件通知

## API 集成

### API 客户端配置

```typescript
// src/services/api.ts
const apiClient = new ApiClient({
  baseURL: process.env.REACT_APP_API_URL,
  timeout: 10000,
});
```

### 服务层示例

```typescript
// src/services/agent.ts
export class AgentService {
  static async getAgents(params?: GetAgentsParams): Promise<ApiResponse<Agent[]>> {
    return apiClient.get('/agents', { params });
  }
  
  static async createAgent(agent: CreateAgentRequest): Promise<ApiResponse<Agent>> {
    return apiClient.post('/agents', agent);
  }
}
```

### 状态管理

```typescript
// src/store/slices/agentSlice.ts
export const fetchAgents = createAsyncThunk(
  'agents/fetchAgents',
  async (params?: GetAgentsParams) => {
    return await AgentService.getAgents(params);
  }
);
```

## 组件开发

### 组件结构

```typescript
// 组件基本结构
import React from 'react';
import { Button, Card } from 'antd';

interface ExampleComponentProps {
  title: string;
  onAction: () => void;
}

const ExampleComponent: React.FC<ExampleComponentProps> = ({ title, onAction }) => {
  return (
    <Card title={title}>
      <Button onClick={onAction}>Action</Button>
    </Card>
  );
};

export default ExampleComponent;
```

### 样式指南

使用 Tailwind CSS 类名结合 Ant Design 组件：

```typescript
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow-sm">
  <Card className="w-full">
    <Button type="primary" className="bg-blue-500 hover:bg-blue-600">
      Submit
    </Button>
  </Card>
</div>
```

## 构建和部署

### 开发构建
```bash
npm run build
# 输出到 dist/ 目录
```

### Docker 部署

1. **构建镜像**
   ```bash
   docker build -t multiagent-frontend .
   ```

2. **运行容器**
   ```bash
   docker run -p 3000:80 multiagent-frontend
   ```

3. **使用 Docker Compose**
   ```bash
   # 在项目根目录
   docker-compose up frontend
   ```

### 生产部署

1. **环境变量配置**
   ```bash
   # 生产环境变量
   REACT_APP_API_URL=https://api.yourdomain.com
   REACT_APP_WS_URL=wss://api.yourdomain.com
   NODE_ENV=production
   ```

2. **Nginx 配置**
   - 静态文件服务
   - API 代理
   - WebSocket 支持
   - Gzip 压缩

## 测试

### 单元测试
```bash
npm run test
```

### E2E 测试
```bash
npm run test:e2e
```

### 代码覆盖率
```bash
npm run test:coverage
```

## 性能优化

### 代码分割
- 使用 React.lazy() 进行组件懒加载
- 路由级别的代码分割

### 缓存策略
- API 响应缓存
- 静态资源缓存
- 浏览器缓存

### 包大小优化
- Tree shaking
- 动态导入
- 依赖分析

## 故障排除

### 常见问题

1. **API 连接失败**
   - 检查 `.env` 文件中的 API URL 配置
   - 确认后端服务是否正常运行

2. **WebSocket 连接失败**
   - 检查 WebSocket URL 配置
   - 确认网络代理设置

3. **构建失败**
   - 清除 node_modules 并重新安装
   - 检查 Node.js 版本兼容性

### 调试工具

- React Developer Tools
- Redux DevTools
- 浏览器开发者工具
- Network 面板查看 API 请求

## 贡献指南

### 代码规范
- 使用 TypeScript 进行类型安全
- 遵循 ESLint 和 Prettier 配置
- 组件使用 PascalCase 命名
- 文件使用 camelCase 命名

### Git 工作流
1. 创建功能分支
2. 提交 PR
3. 代码审查
4. 合并到主分支

### 提交信息格式
```
feat: 添加 Agent 详情页面
fix: 修复 WebSocket 连接问题
docs: 更新 API 文档
style: 优化组件样式
refactor: 重构状态管理逻辑
test: 添加单元测试
```

## 扩展开发

### 添加新页面
1. 在 `src/pages/` 创建新组件
2. 在 `App.tsx` 中添加路由
3. 更新导航菜单

### 添加新 API 服务
1. 在 `src/services/` 创建服务类
2. 定义 TypeScript 类型
3. 创建 Redux slice（如需要）

### 主题定制
- 修改 `tailwind.config.js`
- 调整 Ant Design 主题配置
- 添加自定义 CSS 变量

## 相关资源

- [React 官方文档](https://react.dev/)
- [TypeScript 文档](https://www.typescriptlang.org/)
- [Ant Design 文档](https://ant.design/)
- [Tailwind CSS 文档](https://tailwindcss.com/)
- [Vite 文档](https://vitejs.dev/)
- [Redux Toolkit 文档](https://redux-toolkit.js.org/)