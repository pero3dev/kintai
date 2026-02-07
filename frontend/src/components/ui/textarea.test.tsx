import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Textarea } from './textarea';

describe('Textarea', () => {
  it('renders textarea element', () => {
    render(<Textarea data-testid="textarea" />);
    expect(screen.getByTestId('textarea')).toBeInTheDocument();
  });

  it('handles text input', () => {
    const handleChange = vi.fn();
    render(<Textarea onChange={handleChange} />);

    const textarea = screen.getByRole('textbox');
    fireEvent.change(textarea, { target: { value: 'test value' } });
    expect(handleChange).toHaveBeenCalled();
  });

  it('accepts value prop', () => {
    render(<Textarea value="test value" readOnly />);
    expect(screen.getByRole('textbox')).toHaveValue('test value');
  });

  it('applies base styles', () => {
    render(<Textarea data-testid="textarea" />);
    const textarea = screen.getByTestId('textarea');
    expect(textarea).toHaveClass('flex', 'w-full', 'rounded-md', 'border');
  });

  it('applies min-height style', () => {
    render(<Textarea data-testid="textarea" />);
    expect(screen.getByTestId('textarea')).toHaveClass('min-h-[80px]');
  });

  it('accepts custom className', () => {
    render(<Textarea className="custom-textarea" data-testid="textarea" />);
    expect(screen.getByTestId('textarea')).toHaveClass('custom-textarea');
  });

  it('can be disabled', () => {
    render(<Textarea disabled />);
    expect(screen.getByRole('textbox')).toBeDisabled();
  });

  it('accepts placeholder', () => {
    render(<Textarea placeholder="Enter message" />);
    expect(screen.getByPlaceholderText('Enter message')).toBeInTheDocument();
  });

  it('forwards ref', () => {
    const ref = { current: null };
    render(<Textarea ref={ref} />);
    expect(ref.current).toBeInstanceOf(HTMLTextAreaElement);
  });

  it('passes through additional props', () => {
    render(<Textarea aria-label="Message" name="message" rows={5} data-testid="textarea" />);
    const textarea = screen.getByRole('textbox', { name: 'Message' });
    expect(textarea).toHaveAttribute('name', 'message');
    expect(textarea).toHaveAttribute('rows', '5');
  });

  it('applies disabled opacity', () => {
    render(<Textarea disabled data-testid="textarea" />);
    expect(screen.getByTestId('textarea')).toHaveClass('disabled:opacity-50');
  });
});
