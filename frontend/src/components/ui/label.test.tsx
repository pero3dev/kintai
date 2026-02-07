import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Label } from './label';

describe('Label', () => {
  it('renders label with text', () => {
    render(<Label>Username</Label>);
    expect(screen.getByText('Username')).toBeInTheDocument();
  });

  it('applies base styles', () => {
    render(<Label data-testid="label">Label</Label>);
    const label = screen.getByTestId('label');
    expect(label).toHaveClass('text-sm', 'font-medium', 'leading-none');
  });

  it('accepts custom className', () => {
    render(<Label className="custom-label" data-testid="label">Label</Label>);
    expect(screen.getByTestId('label')).toHaveClass('custom-label');
  });

  it('forwards ref', () => {
    const ref = { current: null };
    render(<Label ref={ref}>Label</Label>);
    expect(ref.current).toBeInstanceOf(HTMLLabelElement);
  });

  it('accepts htmlFor attribute', () => {
    render(<Label htmlFor="email">Email</Label>);
    expect(screen.getByText('Email')).toHaveAttribute('for', 'email');
  });

  it('renders as label element', () => {
    render(<Label>Test Label</Label>);
    expect(screen.getByText('Test Label').tagName).toBe('LABEL');
  });

  it('passes through additional props', () => {
    render(<Label data-testid="label" id="my-label">Label</Label>);
    expect(screen.getByTestId('label')).toHaveAttribute('id', 'my-label');
  });
});
