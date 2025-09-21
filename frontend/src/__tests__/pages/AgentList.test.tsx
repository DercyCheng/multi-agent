import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import AgentList from '../../pages/Agents/AgentList';

describe('AgentList Component', () => {
  test('renders agent list title', () => {
    render(<AgentList />);
    expect(screen.getByText('Agent管理')).toBeTruthy();
  });

  test('renders new agent button', () => {
    render(<AgentList />);
    expect(screen.getByText('新建Agent')).toBeTruthy();
  });

  test('renders agent table headers', () => {
    render(<AgentList />);
    expect(screen.getByText('名称')).toBeTruthy();
    expect(screen.getByText('类型')).toBeTruthy();
    expect(screen.getByText('状态')).toBeTruthy();
    expect(screen.getByText('操作')).toBeTruthy();
  });

  test('displays mock agent data', () => {
    render(<AgentList />);
    expect(screen.getByText('GPT-4 Agent')).toBeTruthy();
    expect(screen.getByText('LLM')).toBeTruthy();
  });
});