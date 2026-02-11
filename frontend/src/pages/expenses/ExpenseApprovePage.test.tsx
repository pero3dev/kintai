import { getApiMocks, resetHarness, setQueryData, setUserRole } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseApprovePage } from './ExpenseApprovePage';

describe('ExpenseApprovePage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('approves a pending expense', async () => {
    setUserRole('manager');
    setQueryData(['expenses', 'pending', 1, 10], {
      data: [
        {
          id: 'exp-1',
          title: 'Taxi',
          user_name: 'Taro Yamada',
          amount: 1800,
          category: 'transportation',
          expense_date: '2026-02-10',
        },
      ],
      total: 1,
    });

    render(<ExpenseApprovePage />);

    fireEvent.click(screen.getAllByRole('button', { name: 'common.approve' })[0]);

    await waitFor(() => {
      expect(getApiMocks().expenses.approve).toHaveBeenCalledWith('exp-1', { status: 'approved' });
    });
  });
});
