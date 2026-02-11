import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ComponentProps } from 'react';
import { Pagination } from './Pagination';

function renderPagination(
  overrides: Partial<ComponentProps<typeof Pagination>> = {},
) {
  const onPageChange = vi.fn();
  const onPageSizeChange = vi.fn();

  const result = render(
    <Pagination
      currentPage={1}
      totalPages={10}
      totalItems={100}
      pageSize={10}
      onPageChange={onPageChange}
      onPageSizeChange={onPageSizeChange}
      {...overrides}
    />,
  );

  return { ...result, onPageChange, onPageSizeChange };
}

describe('Pagination', () => {
  it('returns null when pagination is not needed', () => {
    const onPageChange = vi.fn();
    const { container } = render(
      <Pagination
        currentPage={1}
        totalPages={1}
        totalItems={10}
        pageSize={10}
        onPageChange={onPageChange}
      />,
    );

    expect(container.firstChild).toBeNull();
  });

  it('renders page size selector and handles page size change', () => {
    const { onPageSizeChange } = renderPagination({
      pageSize: 20,
      pageSizeOptions: [10, 20, 50],
    });

    fireEvent.change(screen.getByRole('combobox'), { target: { value: '50' } });

    expect(onPageSizeChange).toHaveBeenCalledWith(50);
  });

  it('does not render page size selector when onPageSizeChange is not provided', () => {
    renderPagination({ onPageSizeChange: undefined });

    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
  });

  it('shows empty-state text block when totalItems is 0', () => {
    const { container } = renderPagination({
      totalItems: 0,
      totalPages: 2,
      onPageSizeChange: undefined,
    });

    expect(container.querySelectorAll('span.font-medium')).toHaveLength(0);
  });

  it('renders all pages when totalPages is 5 or less', () => {
    renderPagination({
      currentPage: 3,
      totalPages: 5,
      totalItems: 50,
      onPageSizeChange: undefined,
    });

    for (const page of ['1', '2', '3', '4', '5']) {
      expect(screen.getByRole('button', { name: page })).toBeInTheDocument();
    }
    expect(screen.queryByText('...')).not.toBeInTheDocument();
  });

  it('renders leading range with ellipsis when currentPage is near the start', () => {
    renderPagination({
      currentPage: 2,
      totalPages: 10,
      totalItems: 100,
      onPageSizeChange: undefined,
    });

    expect(screen.getByRole('button', { name: '4' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '10' })).toBeInTheDocument();
    expect(screen.getByText('...')).toBeInTheDocument();
  });

  it('renders trailing range with ellipsis when currentPage is near the end', () => {
    renderPagination({
      currentPage: 9,
      totalPages: 10,
      totalItems: 100,
      onPageSizeChange: undefined,
    });

    expect(screen.getByRole('button', { name: '7' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '10' })).toBeInTheDocument();
    expect(screen.getByText('...')).toBeInTheDocument();
  });

  it('renders middle range with two ellipses when currentPage is in the middle', () => {
    renderPagination({
      currentPage: 5,
      totalPages: 10,
      totalItems: 100,
      onPageSizeChange: undefined,
    });

    expect(screen.getByRole('button', { name: '4' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '5' })).toHaveClass(
      'gradient-primary',
    );
    expect(screen.getByRole('button', { name: '6' })).toBeInTheDocument();
    expect(screen.getAllByText('...')).toHaveLength(2);
  });

  it('handles page navigation button clicks and disabled state on first page', () => {
    const { container, onPageChange } = renderPagination({
      currentPage: 1,
      totalPages: 10,
      totalItems: 100,
      onPageSizeChange: undefined,
    });
    const navButtons = container.querySelectorAll('button[title]');

    expect(navButtons).toHaveLength(4);
    expect(navButtons[0]).toBeDisabled();
    expect(navButtons[1]).toBeDisabled();
    expect(navButtons[2]).not.toBeDisabled();
    expect(navButtons[3]).not.toBeDisabled();

    fireEvent.click(navButtons[2]);
    fireEvent.click(navButtons[3]);
    fireEvent.click(screen.getByRole('button', { name: '3' }));

    expect(onPageChange).toHaveBeenCalledWith(2);
    expect(onPageChange).toHaveBeenCalledWith(10);
    expect(onPageChange).toHaveBeenCalledWith(3);
  });

  it('disables next and last buttons on final page', () => {
    const { container } = renderPagination({
      currentPage: 10,
      totalPages: 10,
      totalItems: 100,
      onPageSizeChange: undefined,
    });
    const navButtons = container.querySelectorAll('button[title]');

    expect(navButtons[0]).not.toBeDisabled();
    expect(navButtons[1]).not.toBeDisabled();
    expect(navButtons[2]).toBeDisabled();
    expect(navButtons[3]).toBeDisabled();
  });
});
