import {
  getApiMocks,
  resetHarness,
  setQueryData,
  setUserRole,
} from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HolidaysPage } from './HolidaysPage';

describe('HolidaysPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('opens create form and submits holiday data', async () => {
    setUserRole('admin');
    const now = new Date();
    const year = now.getFullYear();
    const month = now.getMonth() + 1;

    setQueryData(['holidays', year], []);
    setQueryData(['holidays', 'calendar', year, month], []);
    setQueryData(['holidays', 'working-days', year, month], {
      working_days: 20,
      holiday_count: 1,
      weekends: 8,
      total_days: 29,
    });

    render(<HolidaysPage />);

    fireEvent.click(screen.getByRole('button', { name: 'holidays.add' }));
    fireEvent.change(document.querySelector('input[name="date"]') as HTMLInputElement, {
      target: { value: '2026-02-11' },
    });
    fireEvent.change(document.querySelector('input[name="name"]') as HTMLInputElement, {
      target: { value: 'Foundation Day' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));

    await waitFor(() => {
      expect(getApiMocks().holidays.create).toHaveBeenCalledWith(
        expect.objectContaining({
          date: '2026-02-11',
          name: 'Foundation Day',
          holiday_type: 'national',
        }),
      );
    });
  });
});
