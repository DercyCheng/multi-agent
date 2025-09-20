import { configureStore } from '@reduxjs/toolkit';
import { agentSlice } from './slices/agentSlice';
import { workflowSlice } from './slices/workflowSlice';
import { authSlice } from './slices/authSlice';
import { systemSlice } from './slices/systemSlice';

export const store = configureStore({
    reducer: {
        auth: authSlice.reducer,
        agents: agentSlice.reducer,
        workflows: workflowSlice.reducer,
        system: systemSlice.reducer,
    },
    middleware: (getDefaultMiddleware: any) =>
        getDefaultMiddleware({
            serializableCheck: {
                ignoredActions: ['persist/PERSIST'],
            },
        }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;