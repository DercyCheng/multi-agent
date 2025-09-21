import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';

// 简单的测试组件
const TestComponent = () => {
  return <div>Hello Test</div>;
};

describe('Simple Test', () => {
  test('renders test component', () => {
    render(<TestComponent />);
    expect(screen.getByText('Hello Test')).toBeInTheDocument();
  });

  test('basic math test', () => {
    expect(2 + 2).toBe(4);
  });
});