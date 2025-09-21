import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import MainLayout from '../../components/Layout/MainLayout';

// Mock react-router-dom hooks
const mockNavigate = jest.fn();
const mockLocation = { pathname: '/dashboard' };

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  useLocation: () => mockLocation,
  Outlet: () => <div data-testid="outlet">Content Area</div>,
}));

const MockedMainLayout = () => (
  <BrowserRouter>
    <MainLayout />
  </BrowserRouter>
);

describe('MainLayout Component', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  test('renders main layout structure', () => {
    render(<MockedMainLayout />);
    
    // Check if main elements are present
    expect(screen.getByText('Multi-Agent 管理平台')).toBeTruthy();
    expect(screen.getByText('仪表板')).toBeTruthy();
    expect(screen.getByText('Agent管理')).toBeTruthy();
    expect(screen.getByText('工作流')).toBeTruthy();
    expect(screen.getByText('管理员')).toBeTruthy();
  });

  test('sidebar shows correct brand text', () => {
    render(<MockedMainLayout />);
    
    // Initially should show full text
    expect(screen.getByText('Multi-Agent')).toBeTruthy();
  });

  test('sidebar can be collapsed', () => {
    render(<MockedMainLayout />);
    
    // Find the collapse button by its icon
    const collapseButtons = screen.getAllByRole('button');
    const collapseButton = collapseButtons.find(button => 
      button.querySelector('.anticon-menu-fold')
    );
    
    expect(collapseButton).toBeTruthy();
    
    if (collapseButton) {
      fireEvent.click(collapseButton);
      // After collapse, should show abbreviated text
      expect(screen.getByText('MA')).toBeTruthy();
    }
  });

  test('navigation menu items are present', () => {
    render(<MockedMainLayout />);
    
    expect(screen.getByText('仪表板')).toBeTruthy();
    expect(screen.getByText('Agent管理')).toBeTruthy();
    expect(screen.getByText('工作流')).toBeTruthy();
    expect(screen.getByText('特性开关')).toBeTruthy();
    expect(screen.getByText('配置中心')).toBeTruthy();
  });

  test('user dropdown is displayed', () => {
    render(<MockedMainLayout />);
    
    // Find user dropdown trigger
    const userDropdown = screen.getByText('管理员');
    expect(userDropdown).toBeTruthy();
  });

  test('notification badge is displayed', () => {
    render(<MockedMainLayout />);
    
    // Check if notification bell icon is present
    const bellIcons = document.querySelectorAll('.anticon-bell');
    expect(bellIcons.length).toBeGreaterThan(0);
  });

  test('content area renders outlet', () => {
    render(<MockedMainLayout />);
    
    expect(screen.getByTestId('outlet')).toBeTruthy();
    expect(screen.getByText('Content Area')).toBeTruthy();
  });
});