import type { Agent, PaginatedResponse, ApiResponse } from '../types';

// Mock service for testing
export class AgentService {
  static async getAgents(params?: { 
    page?: number; 
    pageSize?: number; 
    status?: string; 
    type?: string 
  }): Promise<PaginatedResponse<Agent>> {
    // Mock implementation
    return {
      data: {
        items: [],
        total: 0,
        page: params?.page || 1,
        pageSize: params?.pageSize || 10,
        totalPages: 0,
      },
    };
  }

  static async getAgent(id: string): Promise<ApiResponse<Agent>> {
    // Mock implementation
    return {
      data: {
        id,
        name: 'Mock Agent',
        type: 'LLM',
        status: 'active',
        capabilities: [],
        configuration: {},
        metrics: {
          requestCount: 0,
          successRate: 0,
          averageResponseTime: 0,
          lastActive: new Date().toISOString(),
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      success: true,
    };
  }

  static async createAgent(agent: Partial<Agent>): Promise<ApiResponse<Agent>> {
    // Mock implementation
    return {
      data: {
        id: Date.now().toString(),
        name: agent.name || 'New Agent',
        type: agent.type || 'LLM',
        status: 'inactive',
        capabilities: agent.capabilities || [],
        configuration: agent.configuration || {},
        metrics: {
          requestCount: 0,
          successRate: 0,
          averageResponseTime: 0,
          lastActive: new Date().toISOString(),
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      success: true,
    };
  }

  static async updateAgent(id: string, agent: Partial<Agent>): Promise<ApiResponse<Agent>> {
    // Mock implementation
    return {
      data: {
        id,
        name: agent.name || 'Updated Agent',
        type: agent.type || 'LLM',
        status: agent.status || 'inactive',
        capabilities: agent.capabilities || [],
        configuration: agent.configuration || {},
        metrics: {
          requestCount: 0,
          successRate: 0,
          averageResponseTime: 0,
          lastActive: new Date().toISOString(),
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      success: true,
    };
  }

  static async deleteAgent(id: string): Promise<void> {
    // Mock implementation
    return Promise.resolve();
  }

  static async startAgent(id: string): Promise<ApiResponse<Agent>> {
    // Mock implementation
    return this.updateAgent(id, { status: 'active' });
  }

  static async stopAgent(id: string): Promise<ApiResponse<Agent>> {
    // Mock implementation
    return this.updateAgent(id, { status: 'inactive' });
  }
}