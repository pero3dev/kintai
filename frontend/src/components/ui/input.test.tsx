import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Input } from './input';

describe('Input', () => {
  it('renders input element', () => {
    render(<Input data-testid="input" />);
    expect(screen.getByTestId('input')).toBeInTheDocument();
  });

  it('handles text input', () => {
    const handleChange = vi.fn();
    render(<Input onChange={handleChange} />);
    
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'test value' } });
    expect(handleChange).toHaveBeenCalled();
  });

  it('accepts value prop', () => {
    render(<Input value="test value" readOnly />);
    expect(screen.getByRole('textbox')).toHaveValue('test value');
  });

  it('applies base styles', () => {
    render(<Input data-testid="input" />);
    const input = screen.getByTestId('input');
    expect(input).toHaveClass('flex', 'h-10', 'w-full', 'rounded-md', 'border');
  });

  it('accepts custom className', () => {
    render(<Input className="custom-input" data-testid="input" />);
    expect(screen.getByTestId('input')).toHaveClass('custom-input');
  });

  it('can be disabled', () => {
    render(<Input disabled />);
    expect(screen.getByRole('textbox')).toBeDisabled();
  });

  it('supports different types', () => {
    render(<Input type="email" data-testid="email" />);
    expect(screen.getByTestId('email')).toHaveAttribute('type', 'email');
  });

  it('supports type password', () => {
    render(<Input type="password" data-testid="password" />);
    expect(screen.getByTestId('password')).toHaveAttribute('type', 'password');
  });

  it('supports type number', () => {
    render(<Input type="number" role="spinbutton" data-testid="number" />);
    expect(screen.getByTestId('number')).toHaveAttribute('type', 'number');
  });

  it('accepts placeholder', () => {
    render(<Input placeholder="Enter text" />);
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
  });

  it('forwards ref', () => {
    const ref = { current: null };
    render(<Input ref={ref} />);
    expect(ref.current).toBeInstanceOf(HTMLInputElement);
  });

  it('passes through additional props', () => {
    render(<Input aria-label="Email address" name="email" />);
    const input = screen.getByRole('textbox', { name: 'Email address' });
    expect(input).toHaveAttribute('name', 'email');
  });

  it('applies disabled opacity', () => {
    render(<Input disabled data-testid="input" />);
    expect(screen.getByTestId('input')).toHaveClass('disabled:opacity-50');
  });
});
