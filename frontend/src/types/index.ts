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

// 特性开关相关类型
export interface FeatureFlag {
    id: string;
    name: string;
    description: string;
    enabled: boolean;
    rules: FeatureFlagRule[];
    rollout_percentage: number;
    environments: string[];
    tenant_id: string;
    created_by: string;
    created_at: string;
    updated_at: string;
}

export interface FeatureFlagRule {
    id: string;
    conditions: FeatureFlagCondition[];
    enabled: boolean;
    rollout_percentage: number;
}

export interface FeatureFlagCondition {
    property: string;
    operator: 'equals' | 'not_equals' | 'contains' | 'in' | 'not_in' | 'greater_than' | 'less_than';
    value: any;
}

// 配置中心相关类型
export interface Configuration {
    id: string;
    key: string;
    value: string;
    type: 'string' | 'number' | 'boolean' | 'json';
    environment: string;
    description: string;
    tenant_id: string;
    created_by: string;
    created_at: string;
    updated_at: string;
    version?: string;
}

// 定时任务相关类型
export interface CronJob {
    id: string;
    name: string;
    description: string;
    schedule: string;
    command: string;
    enabled: boolean;
    timeout: number;
    retries: number;
    tenant_id: string;
    created_by: string;
    created_at: string;
    updated_at: string;
    last_run?: string;
    next_run?: string;
    success_rate?: number;
    total_runs?: number;
}

export interface CronJobExecution {
    id: string;
    job_id: string;
    status: 'scheduled' | 'running' | 'completed' | 'failed' | 'timeout';
    started_at: string;
    finished_at?: string;
    duration: number;
    exit_code?: number;
    output: string;
    error?: string;
    attempt: number;
    trigger_type: 'scheduled' | 'manual';
}

// 服务发现相关类型
export interface Service {
    id: string;
    name: string;
    version: string;
    address: string;
    port: number;
    health_check_url: string;
    status: 'healthy' | 'unhealthy' | 'unknown';
    last_seen: string;
    registered_at: string;
    metadata: Record<string, any>;
    tags: string[];
    health_check_interval: number;
    health_check_timeout: number;
    environment: string;
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