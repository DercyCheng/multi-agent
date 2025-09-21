import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { AgentService } from '../../services/agent';
import type { Agent, PaginatedResponse } from '../../types';

interface AgentState {
    agents: Agent[];
    currentAgent: Agent | null;
    loading: boolean;
    error: string | null;
    pagination: {
        total: number;
        page: number;
        pageSize: number;
        totalPages: number;
    };
}

const initialState: AgentState = {
    agents: [],
    currentAgent: null,
    loading: false,
    error: null,
    pagination: {
        total: 0,
        page: 1,
        pageSize: 10,
        totalPages: 0,
    },
};

// 异步操作
export const fetchAgents = createAsyncThunk(
    'agents/fetchAgents',
    async (params?: { page?: number; pageSize?: number; status?: string; type?: string }) => {
        const response = await AgentService.getAgents(params);
        return response;
    }
);

export const fetchAgent = createAsyncThunk(
    'agents/fetchAgent',
    async (id: string) => {
        const response = await AgentService.getAgent(id);
        return response.data;
    }
);

export const createAgent = createAsyncThunk(
    'agents/createAgent',
    async (agent: Partial<Agent>) => {
        const response = await AgentService.createAgent(agent);
        return response.data;
    }
);

export const updateAgent = createAsyncThunk(
    'agents/updateAgent',
    async ({ id, agent }: { id: string; agent: Partial<Agent> }) => {
        const response = await AgentService.updateAgent(id, agent);
        return response.data;
    }
);

export const deleteAgent = createAsyncThunk(
    'agents/deleteAgent',
    async (id: string) => {
        await AgentService.deleteAgent(id);
        return id;
    }
);

export const startAgent = createAsyncThunk(
    'agents/startAgent',
    async (id: string) => {
        const response = await AgentService.startAgent(id);
        return response.data;
    }
);

export const stopAgent = createAsyncThunk(
    'agents/stopAgent',
    async (id: string) => {
        const response = await AgentService.stopAgent(id);
        return response.data;
    }
);

export const agentSlice = createSlice({
    name: 'agents',
    initialState,
    reducers: {
        clearError: (state) => {
            state.error = null;
        },
        clearCurrentAgent: (state) => {
            state.currentAgent = null;
        },
    },
    extraReducers: (builder) => {
        builder
            // fetchAgents
            .addCase(fetchAgents.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchAgents.fulfilled, (state, action) => {
                state.loading = false;
                state.agents = action.payload.data.items;
                state.pagination = {
                    total: action.payload.data.total,
                    page: action.payload.data.page,
                    pageSize: action.payload.data.pageSize,
                    totalPages: action.payload.data.totalPages,
                };
            })
            .addCase(fetchAgents.rejected, (state, action) => {
                state.loading = false;
                state.error = action.error.message || 'Failed to fetch agents';
            })
            // fetchAgent
            .addCase(fetchAgent.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchAgent.fulfilled, (state, action) => {
                state.loading = false;
                state.currentAgent = action.payload;
            })
            .addCase(fetchAgent.rejected, (state, action) => {
                state.loading = false;
                state.error = action.error.message || 'Failed to fetch agent';
            })
            // createAgent
            .addCase(createAgent.fulfilled, (state, action) => {
                state.agents.push(action.payload);
            })
            // updateAgent
            .addCase(updateAgent.fulfilled, (state, action) => {
                const index = state.agents.findIndex(agent => agent.id === action.payload.id);
                if (index !== -1) {
                    state.agents[index] = action.payload;
                }
                if (state.currentAgent?.id === action.payload.id) {
                    state.currentAgent = action.payload;
                }
            })
            // deleteAgent
            .addCase(deleteAgent.fulfilled, (state, action) => {
                state.agents = state.agents.filter(agent => agent.id !== action.payload);
                if (state.currentAgent?.id === action.payload) {
                    state.currentAgent = null;
                }
            })
            // startAgent & stopAgent
            .addCase(startAgent.fulfilled, (state, action) => {
                const index = state.agents.findIndex(agent => agent.id === action.payload.id);
                if (index !== -1) {
                    state.agents[index] = action.payload;
                }
                if (state.currentAgent?.id === action.payload.id) {
                    state.currentAgent = action.payload;
                }
            })
            .addCase(stopAgent.fulfilled, (state, action) => {
                const index = state.agents.findIndex(agent => agent.id === action.payload.id);
                if (index !== -1) {
                    state.agents[index] = action.payload;
                }
                if (state.currentAgent?.id === action.payload.id) {
                    state.currentAgent = action.payload;
                }
            });
    },
});

export const { clearError, clearCurrentAgent } = agentSlice.actions;