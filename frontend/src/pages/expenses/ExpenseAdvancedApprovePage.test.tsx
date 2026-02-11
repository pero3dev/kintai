import { getApiMocks, resetHarness, setQueryData, setUserRole } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseAdvancedApprovePage } from './ExpenseAdvancedApprovePage';

describe('ExpenseAdvancedApprovePage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('runs advanced approve action', async () => {
    setUserRole('manager');
    setQueryData(['expenses', 'pending-advanced', 1, 10], {
      data: [
        {
          id: 'exp-1',
          title: 'Hotel',
          user_name: 'Taro Yamada',
          amount: 12000,
          status: 'pending',
          current_step: 1,
          created_at: '2026-02-10T00:00:00Z',
        },
      ],
      total: 1,
    });
    setQueryData(['expense-approval-flow-config'], {
      steps: [{ name: 'Manager', approver_role: 'manager', auto_approve_below: 0 }],
    });
    setQueryData(['expense-delegates'], { data: [] });

    render(<ExpenseAdvancedApprovePage />);

    fireEvent.click(screen.getAllByRole('button', { name: 'common.approve' })[0]);

    await waitFor(() => {
      expect(getApiMocks().expenses.advancedApprove).toHaveBeenCalledWith('exp-1', { action: 'approve' });
    });
  });
});

