import { getApiMocks, resetHarness, setQueryData } from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { AttendancePage } from './AttendancePage';

describe('AttendancePage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders and submits clock-in action', async () => {
    setQueryData(['attendance', 'today'], {});
    setQueryData(['attendance', 'summary'], undefined);
    setQueryData(['attendance', 'list', 1, 20], { data: [], total_pages: 0, total: 0 });

    render(<AttendancePage />);

    expect(screen.getByText('nav.attendance')).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText('attendance.note'), {
      target: { value: 'start work' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'attendance.clockIn' }));

    await waitFor(() => {
      expect(getApiMocks().attendance.clockIn).toHaveBeenCalledWith({ note: 'start work' });
    });
  });
});
