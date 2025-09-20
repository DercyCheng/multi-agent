import { apiClient } from './api';
import type { Agent, ApiResponse, PaginatedResponse } from '../types';

export class AgentService {
    static async getAgents(params?: {
        page?: number;
        pageSize?: number;
        status?: string;
        type?: string;
    }): Promise<PaginatedResponse<Agent>> {
        const response = await apiClient.get<PaginatedResponse<Agent>['data']>('/agents', {
            params,
        });
        return {
            success: true,
            data: response.data,
            message: response.message,
        };
    }

    static async getAgent(id: string): Promise<ApiResponse<Agent>> {
        return apiClient.get<Agent>(`/agents/${id}`);
    }

    static async createAgent(agent: Partial<Agent>): Promise<ApiResponse<Agent>> {
        return apiClient.post<Agent>('/agents', agent);
    }

    static async updateAgent(id: string, agent: Partial<Agent>): Promise<ApiResponse<Agent>> {
        return apiClient.put<Agent>(`/agents/${id}`, agent);
    }

    static async deleteAgent(id: string): Promise<ApiResponse<void>> {
        return apiClient.delete<void>(`/agents/${id}`);
    }

    static async startAgent(id: string): Promise<ApiResponse<Agent>> {
        return apiClient.post<Agent>(`/agents/${id}/start`);
    }

    static async stopAgent(id: string): Promise<ApiResponse<Agent>> {
        return apiClient.post<Agent>(`/agents/${id}/stop`);
    }

    static async getAgentMetrics(id: string, timeRange?: string): Promise<ApiResponse<any>> {
        return apiClient.get<any>(`/agents/${id}/metrics`, {
            params: { timeRange },
        });
    }

    static async getAgentLogs(id: string, params?: {
        page?: number;
        pageSize?: number;
        level?: string;
    }): Promise<ApiResponse<any>> {
        return apiClient.get<any>(`/agents/${id}/logs`, {
            params,
        });
    }
}