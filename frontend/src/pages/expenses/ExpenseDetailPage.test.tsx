import { getApiMocks, resetHarness, setQueryData, setRouteParams } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseDetailPage } from './ExpenseDetailPage';

describe('ExpenseDetailPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('submits a draft expense', async () => {
    setRouteParams({ expenseId: 'exp-1' });
    setQueryData(['expense', 'exp-1'], {
      id: 'exp-1',
      title: 'Taxi claim',
      status: 'draft',
      created_at: '2026-02-10T09:00:00Z',
      user_name: 'Taro Yamada',
      items: [
        {
          expense_date: '2026-02-10',
          category: 'transportation',
          description: 'Client visit',
          amount: 1800,
        },
      ],
    });
    setQueryData(['expense-comments', 'exp-1'], { data: [] });
    setQueryData(['expense-history', 'exp-1'], { data: [] });

    render(<ExpenseDetailPage />);

    fireEvent.click(screen.getByRole('button', { name: /expenses\.actions\.submit/ }));

    await waitFor(() => {
      expect(getApiMocks().expenses.update).toHaveBeenCalledWith('exp-1', { status: 'pending' });
    });
  });
});
