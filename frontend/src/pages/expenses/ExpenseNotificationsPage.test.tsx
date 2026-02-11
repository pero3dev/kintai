import { getApiMocks, resetHarness, setQueryData } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseNotificationsPage } from './ExpenseNotificationsPage';

describe('ExpenseNotificationsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('marks all notifications as read', async () => {
    setQueryData(['expense-notifications', 'all'], {
      data: [
        {
          id: 'n-1',
          type: 'approved',
          title: 'Approved',
          message: 'Approved message',
          is_read: false,
          created_at: '2026-02-10T00:00:00Z',
        },
      ],
    });
    setQueryData(['expense-reminders'], { data: [] });
    setQueryData(['expense-notification-settings'], {
      on_approved: true,
      on_rejected: true,
      on_comment: true,
      on_reimbursed: true,
      month_end_reminder: true,
      overdue_reminder: true,
      reminder_days_before: 3,
    });

    render(<ExpenseNotificationsPage />);

    fireEvent.click(screen.getByRole('button', { name: /expenses\.notifications\.markAllRead/ }));

    await waitFor(() => {
      expect(getApiMocks().expenses.markAllNotificationsRead).toHaveBeenCalled();
    });
  });
});
