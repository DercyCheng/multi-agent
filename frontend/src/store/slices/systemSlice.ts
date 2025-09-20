import { createSlice } from '@reduxjs/toolkit';
import type { SystemMetrics } from '../../types';

interface SystemState {
    metrics: SystemMetrics | null;
    isConnected: boolean;
    loading: boolean;
    error: string | null;
}

const initialState: SystemState = {
    metrics: null,
    isConnected: false,
    loading: false,
    error: null,
};

export const systemSlice = createSlice({
    name: 'system',
    initialState,
    reducers: {
        setMetrics: (state: any, action: any) => {
            state.metrics = action.payload;
        },
        setConnectionStatus: (state: any, action: any) => {
            state.isConnected = action.payload;
        },
        setLoading: (state: any, action: any) => {
            state.loading = action.payload;
        },
        setError: (state: any, action: any) => {
            state.error = action.payload;
        },
        clearError: (state: any) => {
            state.error = null;
        },
    },
});

export const { setMetrics, setConnectionStatus, setLoading, setError, clearError } = systemSlice.actions;