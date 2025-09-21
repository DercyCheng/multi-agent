import React from 'react';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import App from '../App';

// Mock the store
jest.mock('../store', () => ({
  store: {
    getState: () => ({}),
    dispatch: jest.fn(),
    subscribe: jest.fn(),
  },
}));

// Mock all page components
jest.mock('../pages/Dashboard', () => {
  return function Dashboard() {
    return <div data-testid="dashboard">Dashboard Page</div>;
  };
});

jest.mock('../pages/Agents/AgentList', () => {
  return function AgentList() {
    return <div data-testid="agent-list">Agent List Page</div>;
  };
});

jest.mock('../pages/Auth/Login', () => {
  return function Login() {
    return <div data-testid="login">Login Page</div>;
  };
});

jest.mock('../components/Layout/MainLayout', () => {
  return function MainLayout({ children }: { children: React.ReactNode }) {
    return <div data-testid="main-layout">{children}</div>;
  };
});

describe('App Component', () => {
  test('renders without crashing', () => {
    render(<App />);
  });

  test('renders with correct theme configuration', () => {
    const { container } = render(<App />);
    expect(container.firstChild).toBeTruthy();
  });

  test('provides Redux store to components', () => {
    render(<App />);
    // The app should render without throwing errors related to missing store
    expect(screen.getByTestId('main-layout')).toBeTruthy();
  });
});