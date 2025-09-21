import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import Dashboard from '../../pages/Dashboard';

describe('Dashboard Component', () => {
  test('renders dashboard title', () => {
    render(<Dashboard />);
    expect(screen.getByText('仪表板')).toBeTruthy();
  });

  test('renders statistics cards', () => {
    render(<Dashboard />);
    expect(screen.getByText('活跃Agent')).toBeTruthy();
    expect(screen.getByText('运行中工作流')).toBeTruthy();
    expect(screen.getByText('在线用户')).toBeTruthy();
    expect(screen.getByText('系统运行时间')).toBeTruthy();
  });

  test('displays correct statistics values', () => {
    render(<Dashboard />);
    expect(screen.getByText('12')).toBeTruthy();
    expect(screen.getByText('8')).toBeTruthy();
    expect(screen.getByText('24')).toBeTruthy();
    expect(screen.getByText('99.9%')).toBeTruthy();
  });
});