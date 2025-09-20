import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { WorkflowService } from '../../services/workflow';
import type { Workflow, WorkflowExecution } from '../../types';

interface WorkflowState {
    workflows: Workflow[];
    currentWorkflow: Workflow | null;
    executions: WorkflowExecution[];
    currentExecution: WorkflowExecution | null;
    loading: boolean;
    error: string | null;
    pagination: {
        total: number;
        page: number;
        pageSize: number;
        totalPages: number;
    };
}

const initialState: WorkflowState = {
    workflows: [],
    currentWorkflow: null,
    executions: [],
    currentExecution: null,
    loading: false,
    error: null,
    pagination: {
        total: 0,
        page: 1,
        pageSize: 10,
        totalPages: 0,
    },
};

export const fetchWorkflows = createAsyncThunk(
    'workflows/fetchWorkflows',
    async (params?: { page?: number; pageSize?: number; status?: string }) => {
        const response = await WorkflowService.getWorkflows(params);
        return response;
    }
);

export const fetchWorkflow = createAsyncThunk(
    'workflows/fetchWorkflow',
    async (id: string) => {
        const response = await WorkflowService.getWorkflow(id);
        return response.data;
    }
);

export const createWorkflow = createAsyncThunk(
    'workflows/createWorkflow',
    async (workflow: Partial<Workflow>) => {
        const response = await WorkflowService.createWorkflow(workflow);
        return response.data;
    }
);

export const updateWorkflow = createAsyncThunk(
    'workflows/updateWorkflow',
    async ({ id, workflow }: { id: string; workflow: Partial<Workflow> }) => {
        const response = await WorkflowService.updateWorkflow(id, workflow);
        return response.data;
    }
);

export const deleteWorkflow = createAsyncThunk(
    'workflows/deleteWorkflow',
    async (id: string) => {
        await WorkflowService.deleteWorkflow(id);
        return id;
    }
);

export const executeWorkflow = createAsyncThunk(
    'workflows/executeWorkflow',
    async ({ id, input }: { id: string; input?: any }) => {
        const response = await WorkflowService.executeWorkflow(id, input);
        return response.data;
    }
);

export const fetchWorkflowExecutions = createAsyncThunk(
    'workflows/fetchWorkflowExecutions',
    async ({ workflowId, params }: {
        workflowId: string;
        params?: { page?: number; pageSize?: number; status?: string }
    }) => {
        const response = await WorkflowService.getWorkflowExecutions(workflowId, params);
        return response;
    }
);

export const workflowSlice = createSlice({
    name: 'workflows',
    initialState,
    reducers: {
        clearError: (state: any) => {
            state.error = null;
        },
        clearCurrentWorkflow: (state: any) => {
            state.currentWorkflow = null;
        },
        clearCurrentExecution: (state: any) => {
            state.currentExecution = null;
        },
    },
    extraReducers: (builder: any) => {
        builder
            .addCase(fetchWorkflows.pending, (state: any) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchWorkflows.fulfilled, (state: any, action: any) => {
                state.loading = false;
                state.workflows = action.payload.data.items;
                state.pagination = {
                    total: action.payload.data.total,
                    page: action.payload.data.page,
                    pageSize: action.payload.data.pageSize,
                    totalPages: action.payload.data.totalPages,
                };
            })
            .addCase(fetchWorkflows.rejected, (state: any, action: any) => {
                state.loading = false;
                state.error = action.error.message || 'Failed to fetch workflows';
            })
            .addCase(createWorkflow.fulfilled, (state: any, action: any) => {
                state.workflows.push(action.payload);
            })
            .addCase(updateWorkflow.fulfilled, (state: any, action: any) => {
                const index = state.workflows.findIndex((w: Workflow) => w.id === action.payload.id);
                if (index !== -1) {
                    state.workflows[index] = action.payload;
                }
            })
            .addCase(deleteWorkflow.fulfilled, (state: any, action: any) => {
                state.workflows = state.workflows.filter((w: Workflow) => w.id !== action.payload);
            })
            .addCase(executeWorkflow.fulfilled, (state: any, action: any) => {
                state.executions.unshift(action.payload);
            });
    },
});

export const { clearError, clearCurrentWorkflow, clearCurrentExecution } = workflowSlice.actions;