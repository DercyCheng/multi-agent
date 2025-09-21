import { apiClient } from './api';
import type { FeatureFlag, Configuration, CronJob, Service } from '../types';

// Feature Flags API
const featureFlagsApi = {
    // 获取所有特性开关
    getAll: async (params?: { environment?: string; enabled?: boolean }) => {
        return apiClient.get<FeatureFlag[]>('/feature-flags', { params });
    },

    // 获取单个特性开关
    getById: async (id: string) => {
        return apiClient.get<FeatureFlag>(`/feature-flags/${id}`);
    },

    // 创建特性开关
    create: async (data: Omit<FeatureFlag, 'id' | 'created_at' | 'updated_at'>) => {
        return apiClient.post<FeatureFlag>('/feature-flags', data);
    },

    // 更新特性开关
    update: async (id: string, data: Partial<FeatureFlag>) => {
        return apiClient.put<FeatureFlag>(`/feature-flags/${id}`, data);
    },

    // 删除特性开关
    delete: async (id: string) => {
        return apiClient.delete(`/feature-flags/${id}`);
    },

    // 切换开关状态
    toggle: async (id: string, enabled: boolean) => {
        return apiClient.put<FeatureFlag>(`/feature-flags/${id}/toggle`, { enabled });
    },

    // 评估特性开关
    evaluate: async (name: string, context: Record<string, any>) => {
        return apiClient.post<{ enabled: boolean; variant?: string }>(`/feature-flags/evaluate`, {
            name,
            context
        });
    },

    // 获取开关使用统计
    getStats: async (id: string) => {
        return apiClient.get<{
            total_evaluations: number;
            enabled_percentage: number;
            environment_breakdown: Record<string, number>;
        }>(`/feature-flags/${id}/stats`);
    }
};

// Configuration API
const configurationApi = {
    // 获取所有配置
    getAll: async (params?: { environment?: string; key_prefix?: string }) => {
        return apiClient.get<Configuration[]>('/configurations', { params });
    },

    // 获取单个配置
    getByKey: async (key: string, environment?: string) => {
        return apiClient.get<Configuration>(`/configurations/${key}`, {
            params: { environment }
        });
    },

    // 批量获取配置
    getBatch: async (keys: string[], environment?: string) => {
        return apiClient.post<Record<string, any>>('/configurations/batch', {
            keys,
            environment
        });
    },

    // 创建配置
    create: async (data: Omit<Configuration, 'id' | 'created_at' | 'updated_at'>) => {
        return apiClient.post<Configuration>('/configurations', data);
    },

    // 更新配置
    update: async (id: string, data: Partial<Configuration>) => {
        return apiClient.put<Configuration>(`/configurations/${id}`, data);
    },

    // 删除配置
    delete: async (id: string) => {
        return apiClient.delete(`/configurations/${id}`);
    },

    // 获取配置历史版本
    getHistory: async (key: string) => {
        return apiClient.get<Configuration[]>(`/configurations/${key}/history`);
    },

    // 回滚到指定版本
    rollback: async (key: string, version: string) => {
        return apiClient.post<Configuration>(`/configurations/${key}/rollback`, { version });
    }
};

// CronJob API
const cronJobApi = {
    // 获取所有任务
    getAll: async (params?: { enabled?: boolean; environment?: string }) => {
        return apiClient.get<CronJob[]>('/cronjobs', { params });
    },

    // 获取单个任务
    getById: async (id: string) => {
        return apiClient.get<CronJob>(`/cronjobs/${id}`);
    },

    // 创建任务
    create: async (data: Omit<CronJob, 'id' | 'created_at' | 'updated_at'>) => {
        return apiClient.post<CronJob>('/cronjobs', data);
    },

    // 更新任务
    update: async (id: string, data: Partial<CronJob>) => {
        return apiClient.put<CronJob>(`/cronjobs/${id}`, data);
    },

    // 删除任务
    delete: async (id: string) => {
        return apiClient.delete(`/cronjobs/${id}`);
    },

    // 启用/禁用任务
    toggle: async (id: string, enabled: boolean) => {
        return apiClient.put<CronJob>(`/cronjobs/${id}/toggle`, { enabled });
    },

    // 手动触发任务
    trigger: async (id: string) => {
        return apiClient.post<{ execution_id: string }>(`/cronjobs/${id}/trigger`);
    },

    // 获取执行历史
    getExecutions: async (id: string, params?: { limit?: number; offset?: number }) => {
        return apiClient.get<any[]>(`/cronjobs/${id}/executions`, { params });
    },

    // 获取执行日志
    getExecutionLogs: async (executionId: string) => {
        return apiClient.get<{ logs: string; error?: string }>(`/cronjobs/executions/${executionId}/logs`);
    },

    // 停止正在运行的任务
    stop: async (executionId: string) => {
        return apiClient.post(`/cronjobs/executions/${executionId}/stop`);
    }
};

// Service Discovery API
const serviceDiscoveryApi = {
    // 获取所有服务
    getAll: async (params?: { environment?: string; status?: string; tags?: string[] }) => {
        return apiClient.get<Service[]>('/services', { params });
    },

    // 获取单个服务
    getById: async (id: string) => {
        return apiClient.get<Service>(`/services/${id}`);
    },

    // 注册服务
    register: async (data: Omit<Service, 'id' | 'status' | 'last_seen' | 'registered_at'>) => {
        return apiClient.post<Service>('/services/register', data);
    },

    // 注销服务
    deregister: async (id: string) => {
        return apiClient.delete(`/services/${id}`);
    },

    // 更新服务信息
    update: async (id: string, data: Partial<Service>) => {
        return apiClient.put<Service>(`/services/${id}`, data);
    },

    // 健康检查
    healthCheck: async (id: string) => {
        return apiClient.post<{ status: string; latency: number }>(`/services/${id}/health`);
    },

    // 获取服务健康历史
    getHealthHistory: async (id: string, params?: { hours?: number }) => {
        return apiClient.get<any[]>(`/services/${id}/health/history`, { params });
    },

    // 发现服务实例
    discover: async (serviceName: string, strategy?: string) => {
        return apiClient.get<Service>(`/services/discover/${serviceName}`, {
            params: { strategy }
        });
    },

    // 获取负载均衡器
    getLoadBalancers: async () => {
        return apiClient.get<any[]>('/load-balancers');
    },

    // 创建负载均衡器
    createLoadBalancer: async (data: any) => {
        return apiClient.post<any>('/load-balancers', data);
    }
};

// 系统联动 API - 增强不同服务间的协调
const systemIntegrationApi = {
    // 获取系统整体状态
    getSystemStatus: async () => {
        return apiClient.get<{
            feature_flags_count: number;
            configurations_count: number;
            active_cronjobs: number;
            healthy_services: number;
            total_services: number;
            system_health: 'healthy' | 'warning' | 'critical';
        }>('/system/status');
    },

    // 配置依赖关系分析
    analyzeDependencies: async () => {
        return apiClient.get<{
            feature_flag_dependencies: Array<{
                flag_name: string;
                dependent_configs: string[];
                dependent_services: string[];
                dependent_jobs: string[];
            }>;
            service_dependencies: Array<{
                service_name: string;
                depends_on: string[];
                dependents: string[];
            }>;
        }>('/system/dependencies');
    },

    // 批量配置更新 (影响分析)
    batchUpdate: async (updates: {
        feature_flags?: Array<{ id: string; changes: Partial<FeatureFlag> }>;
        configurations?: Array<{ id: string; changes: Partial<Configuration> }>;
        cronjobs?: Array<{ id: string; changes: Partial<CronJob> }>;
    }) => {
        return apiClient.post<{
            applied_changes: number;
            failed_changes: number;
            warnings: string[];
            affected_services: string[];
        }>('/system/batch-update', updates);
    },

    // 环境同步
    syncEnvironments: async (source: string, target: string, types: string[]) => {
        return apiClient.post<{
            synced_items: number;
            conflicts: Array<{ type: string; key: string; reason: string }>;
            warnings: string[];
        }>('/system/sync-environments', {
            source_environment: source,
            target_environment: target,
            sync_types: types
        });
    },

    // 获取变更历史和影响分析
    getChangeHistory: async (params?: {
        hours?: number;
        types?: string[];
        environment?: string;
    }) => {
        return apiClient.get<Array<{
            timestamp: string;
            type: 'feature_flag' | 'configuration' | 'cronjob' | 'service';
            action: 'create' | 'update' | 'delete';
            entity_id: string;
            entity_name: string;
            changes: Record<string, any>;
            user: string;
            impact: {
                affected_services: string[];
                estimated_users: number;
                risk_level: 'low' | 'medium' | 'high';
            };
        }>>('/system/change-history', { params });
    },

    // 实时事件流 (用于WebSocket订阅)
    subscribeToEvents: (callback: (event: {
        type: string;
        entity: string;
        action: string;
        data: any;
        timestamp: string;
    }) => void) => {
        // WebSocket连接逻辑
        const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/events`;
        const ws = new WebSocket(wsUrl);

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                callback(data);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        return () => ws.close();
    }
};

// 导出所有API
export {
    featureFlagsApi,
    configurationApi,
    cronJobApi,
    serviceDiscoveryApi,
    systemIntegrationApi
};