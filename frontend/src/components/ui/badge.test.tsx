import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Badge } from './badge';

describe('Badge', () => {
  it('renders badge with text', () => {
    render(<Badge>Status</Badge>);
    expect(screen.getByText('Status')).toBeInTheDocument();
  });

  it('applies base styles', () => {
    render(<Badge data-testid="badge">Badge</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('inline-flex', 'items-center', 'rounded-full', 'text-xs');
  });

  it('applies default variant styles', () => {
    render(<Badge data-testid="badge">Default</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('bg-primary', 'text-primary-foreground');
  });

  it('applies secondary variant styles', () => {
    render(<Badge variant="secondary" data-testid="badge">Secondary</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('bg-secondary', 'text-secondary-foreground');
  });

  it('applies destructive variant styles', () => {
    render(<Badge variant="destructive" data-testid="badge">Destructive</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('bg-destructive', 'text-destructive-foreground');
  });

  it('applies outline variant styles', () => {
    render(<Badge variant="outline" data-testid="badge">Outline</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('text-foreground');
  });

  it('applies success variant styles', () => {
    render(<Badge variant="success" data-testid="badge">Success</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('bg-green-100', 'text-green-800');
  });

  it('applies warning variant styles', () => {
    render(<Badge variant="warning" data-testid="badge">Warning</Badge>);
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveClass('bg-yellow-100', 'text-yellow-800');
  });

  it('accepts custom className', () => {
    render(<Badge className="custom-badge" data-testid="badge">Custom</Badge>);
    expect(screen.getByTestId('badge')).toHaveClass('custom-badge');
  });

  it('passes through additional props', () => {
    render(<Badge data-testid="badge" aria-label="Status badge">Info</Badge>);
    expect(screen.getByTestId('badge')).toHaveAttribute('aria-label', 'Status badge');
  });
});
