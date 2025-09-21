import { configureStore } from '@reduxjs/toolkit';
import { agentSlice } from '../../store/slices/agentSlice';

describe('agentSlice', () => {
  let store: ReturnType<typeof configureStore>;

  beforeEach(() => {
    store = configureStore({
      reducer: {
        agents: agentSlice.reducer,
      },
    });
  });

  test('should have initial state', () => {
    const state = (store.getState() as any).agents;
    expect(state).toBeDefined();
  });

  test('should handle actions without errors', () => {
    // Test that the slice exists and can be used
    expect(agentSlice.reducer).toBeDefined();
    expect(agentSlice.actions).toBeDefined();
  });
});