import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Select } from './select';

describe('Select', () => {
  it('renders select element', () => {
    render(
      <Select data-testid="select">
        <option value="1">Option 1</option>
      </Select>
    );
    expect(screen.getByTestId('select')).toBeInTheDocument();
  });

  it('renders options', () => {
    render(
      <Select>
        <option value="1">Option 1</option>
        <option value="2">Option 2</option>
      </Select>
    );
    expect(screen.getByRole('combobox')).toBeInTheDocument();
    expect(screen.getByText('Option 1')).toBeInTheDocument();
    expect(screen.getByText('Option 2')).toBeInTheDocument();
  });

  it('handles selection change', () => {
    const handleChange = vi.fn();
    render(
      <Select onChange={handleChange}>
        <option value="1">Option 1</option>
        <option value="2">Option 2</option>
      </Select>
    );

    fireEvent.change(screen.getByRole('combobox'), { target: { value: '2' } });
    expect(handleChange).toHaveBeenCalled();
  });

  it('accepts value prop', () => {
    render(
      <Select value="2" onChange={() => { }}>
        <option value="1">Option 1</option>
        <option value="2">Option 2</option>
      </Select>
    );
    expect(screen.getByRole('combobox')).toHaveValue('2');
  });

  it('applies base styles', () => {
    render(
      <Select data-testid="select">
        <option value="1">Option 1</option>
      </Select>
    );
    const select = screen.getByTestId('select');
    expect(select).toHaveClass('flex', 'h-10', 'w-full', 'rounded-md', 'border');
  });

  it('accepts custom className', () => {
    render(
      <Select className="custom-select" data-testid="select">
        <option value="1">Option 1</option>
      </Select>
    );
    expect(screen.getByTestId('select')).toHaveClass('custom-select');
  });

  it('can be disabled', () => {
    render(
      <Select disabled>
        <option value="1">Option 1</option>
      </Select>
    );
    expect(screen.getByRole('combobox')).toBeDisabled();
  });

  it('forwards ref', () => {
    const ref = { current: null };
    render(
      <Select ref={ref}>
        <option value="1">Option 1</option>
      </Select>
    );
    expect(ref.current).toBeInstanceOf(HTMLSelectElement);
  });

  it('passes through additional props', () => {
    render(
      <Select aria-label="Choose option" name="option" data-testid="select">
        <option value="1">Option 1</option>
      </Select>
    );
    const select = screen.getByRole('combobox', { name: 'Choose option' });
    expect(select).toHaveAttribute('name', 'option');
  });

  it('applies disabled opacity', () => {
    render(
      <Select disabled data-testid="select">
        <option value="1">Option 1</option>
      </Select>
    );
    expect(screen.getByTestId('select')).toHaveClass('disabled:opacity-50');
  });
});
