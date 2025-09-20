// Agent相关类型
export interface Agent {
    id: string;
    name: string;
    type: string;
    status: 'active' | 'inactive' | 'error' | 'pending';
    capabilities: string[];
    configuration: Record<string, any>;
    metrics: {
        requestCount: number;
        successRate: number;
        averageResponseTime: number;
        lastActive: string;
    };
    createdAt: string;
    updatedAt: string;
}

// 工作流相关类型
export interface Workflow {
    id: string;
    name: string;
    description: string;
    status: 'draft' | 'active' | 'paused' | 'completed' | 'failed';
    steps: WorkflowStep[];
    metadata: {
        createdBy: string;
        createdAt: string;
        updatedAt: string;
        version: string;
    };
    metrics: {
        totalRuns: number;
        successfulRuns: number;
        failedRuns: number;
        averageExecutionTime: number;
    };
}

export interface WorkflowStep {
    id: string;
    name: string;
    type: 'agent' | 'condition' | 'parallel' | 'sequential';
    agentId?: string;
    configuration: Record<string, any>;
    dependencies: string[];
    timeout: number;
    retryPolicy: {
        maxRetries: number;
        backoffStrategy: 'linear' | 'exponential';
    };
}

// 执行相关类型
export interface WorkflowExecution {
    id: string;
    workflowId: string;
    status: 'running' | 'completed' | 'failed' | 'cancelled';
    startTime: string;
    endTime?: string;
    steps: StepExecution[];
    metadata: Record<string, any>;
    error?: string;
}

export interface StepExecution {
    stepId: string;
    status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
    startTime?: string;
    endTime?: string;
    result?: any;
    error?: string;
    retryCount: number;
}

// 用户和认证相关类型
export interface User {
    id: string;
    username: string;
    email: string;
    role: 'admin' | 'operator' | 'viewer';
    tenantId: string;
    permissions: string[];
    lastLogin: string;
    createdAt: string;
}

export interface Tenant {
    id: string;
    name: string;
    domain: string;
    status: 'active' | 'suspended';
    settings: {
        maxAgents: number;
        maxWorkflows: number;
        retentionDays: number;
    };
    createdAt: string;
}

// API响应类型
export interface ApiResponse<T> {
    success: boolean;
    data: T;
    message: string;
    code: number;
    timestamp: string;
}

export interface PaginatedResponse<T> {
    success: boolean;
    data: {
        items: T[];
        total: number;
        page: number;
        pageSize: number;
        totalPages: number;
    };
    message: string;
}

// 实时数据类型
export interface SystemMetrics {
    timestamp: string;
    cpu: {
        usage: number;
        cores: number;
    };
    memory: {
        used: number;
        total: number;
        usage: number;
    };
    agents: {
        total: number;
        active: number;
        failed: number;
    };
    workflows: {
        running: number;
        completed: number;
        failed: number;
    };
    throughput: {
        requestsPerSecond: number;
        averageResponseTime: number;
    };
}

// 配置相关类型
export interface AppConfig {
    api: {
        baseUrl: string;
        timeout: number;
    };
    auth: {
        tokenKey: string;
        refreshTokenKey: string;
    };
    websocket: {
        url: string;
        reconnectInterval: number;
    };
    ui: {
        theme: 'light' | 'dark';
        language: string;
        pageSize: number;
    };
}