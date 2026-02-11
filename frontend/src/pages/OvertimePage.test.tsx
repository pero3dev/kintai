import {
  getApiMocks,
  resetHarness,
  setQueryData,
  setUserRole,
} from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { OvertimePage } from './OvertimePage';

describe('OvertimePage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders admin pending requests and approves one', async () => {
    setUserRole('admin');
    setQueryData(['overtime', 'my', 1, 20], { data: [], total_pages: 0, total: 0 });
    setQueryData(['overtime', 'pending', 1], {
      data: [
        {
          id: 'ot-1',
          user: { last_name: 'Yamada', first_name: 'Taro' },
          date: '2026-02-11',
          planned_minutes: 90,
          reason: 'Release task',
        },
      ],
      total_pages: 1,
      total: 1,
    });
    setQueryData(['overtime', 'alerts'], [
      {
        user_id: 'u1',
        user_name: 'Yamada Taro',
        monthly_overtime_hours: 46,
        monthly_limit_hours: 45,
        yearly_overtime_hours: 360,
        yearly_limit_hours: 360,
        is_monthly_exceeded: true,
        is_yearly_exceeded: false,
      },
    ]);

    render(<OvertimePage />);

    expect(screen.getByText('overtime.alerts')).toBeInTheDocument();
    fireEvent.click(screen.getAllByText('common.approve')[0]);

    await waitFor(() => {
      expect(getApiMocks().overtime.approve).toHaveBeenCalledWith('ot-1', { status: 'approved' });
    });
  });
});
