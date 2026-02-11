import { getApiMocks, resetHarness, setQueryData } from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { NotificationsPage } from './NotificationsPage';

describe('NotificationsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('marks notifications as read and deletes one', async () => {
    setQueryData(['notifications', 1, 20], {
      data: [
        {
          id: 'n-1',
          type: 'leave_approved',
          title: 'Leave approved',
          message: 'Your leave has been approved.',
          created_at: '2026-02-11T09:00:00Z',
          is_read: false,
        },
        {
          id: 'n-2',
          type: 'general',
          title: 'FYI',
          message: 'General info.',
          created_at: '2026-02-11T10:00:00Z',
          is_read: true,
        },
      ],
      total_pages: 1,
      total: 2,
    });
    setQueryData(['notifications', 'unread-count'], { unread: 1 });

    render(<NotificationsPage />);

    fireEvent.click(screen.getByRole('button', { name: 'notifications.markAllRead' }));
    fireEvent.click(screen.getByTitle('common.markAsRead'));
    fireEvent.click(screen.getAllByTitle('common.delete')[0]);

    await waitFor(() => {
      expect(getApiMocks().notifications.markAllAsRead).toHaveBeenCalled();
      expect(getApiMocks().notifications.markAsRead).toHaveBeenCalledWith('n-1');
      expect(getApiMocks().notifications.delete).toHaveBeenCalledWith('n-1');
    });
  });
});
