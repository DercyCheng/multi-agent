import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import type { User } from '../../types';

interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    loading: boolean;
    error: string | null;
}

const initialState: AuthState = {
    user: null,
    token: localStorage.getItem('auth_token'),
    isAuthenticated: false,
    loading: false,
    error: null,
};

export const login = createAsyncThunk(
    'auth/login',
    async ({ email, password }: { email: string; password: string }) => {
        // This would be replaced with actual API call
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
        });
        const data = await response.json();
        return data;
    }
);

export const logout = createAsyncThunk(
    'auth/logout',
    async () => {
        localStorage.removeItem('auth_token');
        return true;
    }
);

export const authSlice = createSlice({
    name: 'auth',
    initialState,
    reducers: {
        clearError: (state: any) => {
            state.error = null;
        },
        setCredentials: (state: any, action: any) => {
            const { user, token } = action.payload;
            state.user = user;
            state.token = token;
            state.isAuthenticated = true;
            localStorage.setItem('auth_token', token);
        },
    },
    extraReducers: (builder: any) => {
        builder
            .addCase(login.pending, (state: any) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(login.fulfilled, (state: any, action: any) => {
                state.loading = false;
                state.user = action.payload.user;
                state.token = action.payload.token;
                state.isAuthenticated = true;
                localStorage.setItem('auth_token', action.payload.token);
            })
            .addCase(login.rejected, (state: any, action: any) => {
                state.loading = false;
                state.error = action.error.message || 'Login failed';
            })
            .addCase(logout.fulfilled, (state: any) => {
                state.user = null;
                state.token = null;
                state.isAuthenticated = false;
            });
    },
});

export const { clearError, setCredentials } = authSlice.actions;