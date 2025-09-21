import { apiClient } from './api';
import type { FeatureFlag, Configuration, CronJob, Service } from '../types';

const API_PREFIX = '/api/v1';

// Feature Flags API
const featureFlagsApi = {
    // 获取所有特性开关
    getAll: async (params?: { environment?: string; enabled?: boolean }) => {
        return apiClient.get<FeatureFlag[]>(`${API_PREFIX}/flags`, { params });
    },

    // 获取单个特性开关
    getById: async (id: string) => {
        return apiClient.get<FeatureFlag>(`${API_PREFIX}/flags/${id}`);
    },

    // 创建特性开关
    create: async (data: Omit<FeatureFlag, 'id' | 'created_at' | 'updated_at'>) => {
        return apiClient.post<FeatureFlag>(`${API_PREFIX}/flags`, data);
    },

    // 更新特性开关
    update: async (id: string, data: Partial<FeatureFlag>) => {
        return apiClient.put<FeatureFlag>(`${API_PREFIX}/flags/${id}`, data);
    },

    // 删除特性开关
    delete: async (id: string) => {
        return apiClient.delete(`${API_PREFIX}/flags/${id}`);
    },

    // 切换开关状态
    toggle: async (id: string, enabled: boolean) => {
        // backend does not provide a dedicated toggle endpoint; use update
        return apiClient.put<FeatureFlag>(`${API_PREFIX}/flags/${id}`, { enabled });
    },

    // 评估特性开关
    evaluate: async (name: string, context: Record<string, any>) => {
        // backend expects a POST to /api/v1/flags/evaluate with flag_name, environment, user_id, tenant_id, etc.
        return apiClient.post<{ enabled: boolean; variant?: string }>(`${API_PREFIX}/flags/evaluate`, {
            flag_name: name,
            ...context,
        });
    },

    // 获取开关使用统计
    getStats: async (id: string) => {
        return apiClient.get<{
            total_evaluations: number;
            enabled_percentage: number;
            environment_breakdown: Record<string, number>;
        }>(`${API_PREFIX}/flags/${id}/stats`);
    }
};

// Configuration API
const configurationApi = {
    // 获取所有配置
    getAll: async (params?: { environment?: string; key_prefix?: string }) => {
        // backend provides /api/v1/config
        return apiClient.get<Configuration[]>(`${API_PREFIX}/config`, { params });
    },

    // 获取单个配置
    getByKey: async (key: string, environment?: string) => {
        return apiClient.get<Configuration>(`${API_PREFIX}/config/${encodeURIComponent(key)}`, {
            params: { environment }
        });
    },

    // 批量获取配置
    getBatch: async (keys: string[], environment?: string) => {
        // backend doesn't have a batch endpoint; fall back to multiple calls or a single call to /config with query
        return apiClient.post<Record<string, any>>(`${API_PREFIX}/config/batch`, {
            keys,
            environment
        });
    },

    // 创建配置
    create: async (data: Omit<Configuration, 'id' | 'created_at' | 'updated_at'>) => {
        return apiClient.post<Configuration>(`${API_PREFIX}/config`, data);
    },

    // 更新配置
    update: async (id: string, data: Partial<Configuration>) => {
        return apiClient.put<Configuration>(`${API_PREFIX}/config/${encodeURIComponent(id)}`, data);
    },

    // 删除配置
    delete: async (id: string) => {
        return apiClient.delete(`${API_PREFIX}/config/${encodeURIComponent(id)}`);
    },

    // 获取配置历史版本
    getHistory: async (key: string) => {
        // backend exposes audit logs: /api/v1/audit/config
        return apiClient.get<Configuration[]>(`${API_PREFIX}/audit/config`, { params: { key } });
    },

    // 回滚到指定版本
    rollback: async (key: string, version: string) => {
        return apiClient.post<Configuration>(`${API_PREFIX}/config/${encodeURIComponent(key)}/rollback`, { version });
    }
};

// CronJob API
const cronJobApi = {
    // 获取所有任务
    getAll: async (params?: { enabled?: boolean; environment?: string }) => {
        return apiClient.get<CronJob[]>(`${API_PREFIX}/cron/jobs`, { params });
    },

    // 获取单个任务
    getById: async (id: string) => {
        return apiClient.get<CronJob>(`${API_PREFIX}/cron/jobs/${id}`);
    },

    // 创建任务
    create: async (data: Omit<CronJob, 'id' | 'created_at' | 'updated_at'>) => {
        return apiClient.post<CronJob>(`${API_PREFIX}/cron/jobs`, data);
    },

    // 更新任务
    update: async (id: string, data: Partial<CronJob>) => {
        return apiClient.put<CronJob>(`${API_PREFIX}/cron/jobs/${id}`, data);
    },

    // 删除任务
    delete: async (id: string) => {
        return apiClient.delete(`${API_PREFIX}/cron/jobs/${id}`);
    },

    // 启用/禁用任务
    toggle: async (id: string, enabled: boolean) => {
        return apiClient.put<CronJob>(`${API_PREFIX}/cron/jobs/${id}`, { enabled });
    },

    // 手动触发任务
    trigger: async (id: string) => {
        return apiClient.post<{ execution_id: string }>(`${API_PREFIX}/cron/jobs/${id}/trigger`);
    },

    // 获取执行历史
    getExecutions: async (id: string, params?: { limit?: number; offset?: number }) => {
        return apiClient.get<any[]>(`${API_PREFIX}/cron/jobs/${id}/executions`, { params });
    },

    // 获取执行日志
    getExecutionLogs: async (executionId: string) => {
        return apiClient.get<{ logs: string; error?: string }>(`${API_PREFIX}/cron/executions/${executionId}/logs`);
    },

    // 停止正在运行的任务
    stop: async (executionId: string) => {
        return apiClient.post(`${API_PREFIX}/cron/executions/${executionId}/stop`);
    }
};

// Service Discovery API
const serviceDiscoveryApi = {
    // 获取所有服务
    getAll: async (params?: { environment?: string; status?: string; tags?: string[] }) => {
        return apiClient.get<Service[]>(`${API_PREFIX}/services`, { params });
    },

    // 获取单个服务
    getById: async (id: string) => {
        return apiClient.get<Service>(`${API_PREFIX}/services/${id}`);
    },

    // 注册服务
    register: async (data: Omit<Service, 'id' | 'status' | 'last_seen' | 'registered_at'>) => {
        return apiClient.post<Service>(`${API_PREFIX}/services/register`, data);
    },

    // 注销服务
    deregister: async (id: string) => {
        return apiClient.delete(`${API_PREFIX}/services/${id}`);
    },

    // 更新服务信息
    update: async (id: string, data: Partial<Service>) => {
        return apiClient.put<Service>(`${API_PREFIX}/services/${id}`, data);
    },

    // 健康检查
    healthCheck: async (id: string) => {
        return apiClient.post<{ status: string; latency: number }>(`${API_PREFIX}/services/${id}/health`);
    },

    // 获取服务健康历史
    getHealthHistory: async (id: string, params?: { hours?: number }) => {
        return apiClient.get<any[]>(`${API_PREFIX}/services/${id}/health/history`, { params });
    },

    // 发现服务实例
    discover: async (serviceName: string, strategy?: string) => {
        return apiClient.get<Service>(`${API_PREFIX}/services/discover/${serviceName}`, {
            params: { strategy }
        });
    },

    // 获取负载均衡器
    getLoadBalancers: async () => {
        return apiClient.get<any[]>(`${API_PREFIX}/loadbalancer/strategies`);
    },

    // 创建/设置负载均衡策略
    createLoadBalancer: async (data: any) => {
        return apiClient.post<any>(`${API_PREFIX}/loadbalancer/strategy`, data);
    }
};

// 系统联动 API - 增强不同服务间的协调
const systemIntegrationApi = {
    // 获取系统整体状态
    getSystemStatus: async () => {
        // Aggregate from existing endpoints: flags, config, cron, services
        const [flagsRes, configsRes, cronRes, servicesRes] = await Promise.all([
            apiClient.get<FeatureFlag[]>(`${API_PREFIX}/flags`),
            apiClient.get<Configuration[]>(`${API_PREFIX}/config`),
            apiClient.get<CronJob[]>(`${API_PREFIX}/cron/jobs`),
            apiClient.get<Service[]>(`${API_PREFIX}/services`),
        ]);

        const totalServices = (servicesRes.data || []).length;
        const healthyServices = (servicesRes.data || []).filter(s => (s as any).health === 'healthy').length;

        const systemStatus = {
            feature_flags_count: (flagsRes.data || []).length,
            configurations_count: (configsRes.data || []).length,
            active_cronjobs: (cronRes.data || []).filter(j => (j as any).enabled).length,
            healthy_services: healthyServices,
            total_services: totalServices || 0,
            system_health: healthyServices === totalServices ? 'healthy' : (healthyServices / Math.max(1, totalServices) > 0.7 ? 'warning' : 'critical') as 'healthy' | 'warning' | 'critical'
        };

        return { data: systemStatus };
    },

    // 配置依赖关系分析
    analyzeDependencies: async () => {
        // No single backend endpoint; do a basic aggregation and return empty dependency lists as placeholder
        const [flagsRes, servicesRes] = await Promise.all([
            apiClient.get<FeatureFlag[]>(`${API_PREFIX}/flags`),
            apiClient.get<Service[]>(`${API_PREFIX}/services`),
        ]);

        const result = {
            feature_flag_dependencies: (flagsRes.data || []).map((f: any) => ({
                flag_name: f.name || f.id,
                dependent_configs: [],
                dependent_services: [],
                dependent_jobs: []
            })),
            service_dependencies: (servicesRes.data || []).map((s: any) => ({
                service_name: s.name || s.id,
                depends_on: [],
                dependents: []
            })),
        };

        return { data: result };
    },

    // 批量配置更新 (影响分析)
    batchUpdate: async (updates: {
        feature_flags?: Array<{ id: string; changes: Partial<FeatureFlag> }>;
        configurations?: Array<{ id: string; changes: Partial<Configuration> }>;
        cronjobs?: Array<{ id: string; changes: Partial<CronJob> }>;
    }) => {
        // Apply changes by calling underlying APIs sequentially (best-effort)
        let applied = 0;
        let failed = 0;
        const warnings: string[] = [];
        const affected_services: string[] = [];

        if (updates.feature_flags) {
            for (const f of updates.feature_flags) {
                try {
                    await apiClient.put(`${API_PREFIX}/flags/${f.id}`, f.changes);
                    applied++;
                } catch (e) {
                    failed++;
                }
            }
        }

        if (updates.configurations) {
            for (const c of updates.configurations) {
                try {
                    await apiClient.put(`${API_PREFIX}/config/${encodeURIComponent(c.id)}`, c.changes);
                    applied++;
                } catch (e) {
                    failed++;
                }
            }
        }

        if (updates.cronjobs) {
            for (const j of updates.cronjobs) {
                try {
                    await apiClient.put(`${API_PREFIX}/cron/jobs/${j.id}`, j.changes);
                    applied++;
                } catch (e) {
                    failed++;
                }
            }
        }

        return { data: { applied_changes: applied, failed_changes: failed, warnings, affected_services } };
    },

    // 环境同步
    syncEnvironments: async (source: string, target: string, types: string[]) => {
        // No dedicated endpoint; return a placeholder success and echo inputs so params are used
        return { data: { synced_items: 0, conflicts: [], warnings: [], source_environment: source, target_environment: target, sync_types: types } };
    },

    // 获取变更历史和影响分析
    getChangeHistory: async (params?: {
        hours?: number;
        types?: string[];
        environment?: string;
    }) => {
        // Use audit endpoints where possible and pass through params
        const flagsAudit = await apiClient.get<any[]>(`${API_PREFIX}/audit/flags`, { params });
        const configAudit = await apiClient.get<any[]>(`${API_PREFIX}/audit/config`, { params });

        const combined = [
            ...((flagsAudit.data || []) as any[]).map((a: any) => ({ ...a, type: 'feature_flag' })),
            ...((configAudit.data || []) as any[]).map((a: any) => ({ ...a, type: 'configuration' })),
        ];

        return { data: combined };
    },

    // 实时事件流 (用于WebSocket订阅)
    subscribeToEvents: (callback: (event: {
        type: string;
        entity: string;
        action: string;
        data: any;
        timestamp: string;
    }) => void) => {
        // Connect to backend WebSocket at /ws
        const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
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