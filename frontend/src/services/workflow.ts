import { apiClient } from './api';
import type { Workflow, WorkflowExecution, ApiResponse, PaginatedResponse } from '../types';

export class WorkflowService {
    static async getWorkflows(params?: {
        page?: number;
        pageSize?: number;
        status?: string;
    }): Promise<PaginatedResponse<Workflow>> {
        const response = await apiClient.get<PaginatedResponse<Workflow>['data']>('/workflows', {
            params,
        });
        return {
            success: true,
            data: response.data,
            message: response.message,
        };
    }

    static async getWorkflow(id: string): Promise<ApiResponse<Workflow>> {
        return apiClient.get<Workflow>(`/workflows/${id}`);
    }

    static async createWorkflow(workflow: Partial<Workflow>): Promise<ApiResponse<Workflow>> {
        return apiClient.post<Workflow>('/workflows', workflow);
    }

    static async updateWorkflow(id: string, workflow: Partial<Workflow>): Promise<ApiResponse<Workflow>> {
        return apiClient.put<Workflow>(`/workflows/${id}`, workflow);
    }

    static async deleteWorkflow(id: string): Promise<ApiResponse<void>> {
        return apiClient.delete<void>(`/workflows/${id}`);
    }

    static async executeWorkflow(id: string, input?: any): Promise<ApiResponse<WorkflowExecution>> {
        return apiClient.post<WorkflowExecution>(`/workflows/${id}/execute`, { input });
    }

    static async getWorkflowExecutions(workflowId: string, params?: {
        page?: number;
        pageSize?: number;
        status?: string;
    }): Promise<PaginatedResponse<WorkflowExecution>> {
        const response = await apiClient.get<PaginatedResponse<WorkflowExecution>['data']>(
            `/workflows/${workflowId}/executions`,
            { params }
        );
        return {
            success: true,
            data: response.data,
            message: response.message,
        };
    }

    static async getWorkflowExecution(workflowId: string, executionId: string): Promise<ApiResponse<WorkflowExecution>> {
        return apiClient.get<WorkflowExecution>(`/workflows/${workflowId}/executions/${executionId}`);
    }

    static async cancelWorkflowExecution(workflowId: string, executionId: string): Promise<ApiResponse<void>> {
        return apiClient.post<void>(`/workflows/${workflowId}/executions/${executionId}/cancel`);
    }

    static async getWorkflowMetrics(id: string, timeRange?: string): Promise<ApiResponse<any>> {
        return apiClient.get<any>(`/workflows/${id}/metrics`, {
            params: { timeRange },
        });
    }
}