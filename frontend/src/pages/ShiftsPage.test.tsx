import {
  getWeekRange,
  resetHarness,
  setQueryData,
  setUserRole,
} from './__tests__/testHarness';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ShiftsPage } from './ShiftsPage';

describe('ShiftsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders weekly shift board for admin/manager view', () => {
    setUserRole('manager');
    const { startDate, endDate } = getWeekRange(new Date());
    setQueryData(['shifts', startDate, endDate], []);
    setQueryData(['users'], {
      data: [
        { id: 'u2', first_name: 'Hanako', last_name: 'Suzuki', role: 'employee' },
      ],
    });

    render(<ShiftsPage />);

    expect(screen.getByText('shifts.title')).toBeInTheDocument();
    fireEvent.click(screen.getByText('common.thisWeek'));
    expect(screen.getByText('shifts.title')).toBeInTheDocument();
  });
});
