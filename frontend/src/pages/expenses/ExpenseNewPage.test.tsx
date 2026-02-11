import { getApiMocks, resetHarness, setQueryData } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseNewPage } from './ExpenseNewPage';

describe('ExpenseNewPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('creates a new expense on submit', async () => {
    setQueryData(['expense-templates'], { data: [] });

    const { container } = render(<ExpenseNewPage />);

    fireEvent.change(screen.getByPlaceholderText('expenses.placeholders.title'), {
      target: { value: 'Taxi claim' },
    });

    const dateInput = container.querySelector('input[type="date"]');
    const categorySelect = container.querySelector('select');
    const amountInput = container.querySelector('input[type="number"]');
    if (!dateInput || !categorySelect || !amountInput) {
      throw new Error('Required form controls were not found');
    }

    fireEvent.change(dateInput, { target: { value: '2026-02-10' } });
    fireEvent.change(categorySelect, { target: { value: 'transportation' } });
    fireEvent.change(screen.getByPlaceholderText('expenses.placeholders.description'), {
      target: { value: 'Client visit' },
    });
    fireEvent.change(amountInput, { target: { value: '1800' } });

    fireEvent.click(screen.getByRole('button', { name: /expenses\.actions\.submit/ }));

    await waitFor(() => {
      expect(getApiMocks().expenses.create).toHaveBeenCalledWith(
        expect.objectContaining({
          title: 'Taxi claim',
          status: 'pending',
        }),
      );
    });
  });
});
