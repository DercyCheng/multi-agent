export interface Agent {
  id: string;
  name: string;
  type: 'LLM' | 'Tool' | 'Code' | 'Data';
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

export interface Workflow {
  id: string;
  name: string;
  description: string;
  status: 'running' | 'stopped' | 'error' | 'completed' | 'active' | 'draft' | 'paused';
  agents: string[];
  createdAt: string;
  updatedAt: string;
  metadata?: {
    version: string;
    createdBy?: string;
    createdAt: string;
    updatedAt: string;
  };
  metrics?: {
    totalRuns: number;
    successfulRuns?: number;
    failedRuns?: number;
    averageExecutionTime?: number;
  };
  steps?: WorkflowStep[];
}

export interface WorkflowStep {
  id: string;
  name: string;
  type: string;
  agentId: string;
  configuration: Record<string, any>;
  order: number;
  dependencies?: string[];
  timeout?: number;
  retryPolicy?: {
    maxRetries: number;
    backoffStrategy: string;
  };
}

export interface WorkflowExecution {
  id: string;
  workflowId: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  startTime: string;
  endTime?: string;
  steps: WorkflowExecutionStep[];
  result?: any;
  error?: string;
  metadata?: Record<string, any>;
}

export interface WorkflowExecutionStep {
  stepId: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  startTime?: string;
  endTime?: string;
  result?: any;
  error?: string;
  retryCount?: number;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'user';
  createdAt: string;
}

export interface SystemMetrics {
  cpu: number;
  memory: number;
  disk: number;
  network: number;
}

export interface FeatureFlag {
  id: string;
  name: string;
  key: string;
  enabled: boolean;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface Configuration {
  id: string;
  key: string;
  value: any;
  type: 'string' | 'number' | 'boolean' | 'json';
  description: string;
  category: string;
  createdAt: string;
  updatedAt: string;
}

export interface CronJob {
  id: string;
  name: string;
  schedule: string;
  command: string;
  enabled: boolean;
  lastRun?: string;
  nextRun: string;
  status: 'active' | 'inactive' | 'error';
  createdAt: string;
  updatedAt: string;
}

export interface Service {
  id: string;
  name: string;
  type: string;
  status: 'healthy' | 'unhealthy' | 'unknown';
  endpoint: string;
  version: string;
  lastHealthCheck: string;
  metadata: Record<string, any>;
}

export interface PaginatedResponse<T> {
  data: {
    items: T[];
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
  };
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}