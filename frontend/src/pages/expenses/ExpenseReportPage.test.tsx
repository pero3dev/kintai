import { getApiMocks, resetHarness, setQueryData } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { endOfMonth, format, startOfMonth } from 'date-fns';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ExpenseReportPage } from './ExpenseReportPage';

describe('ExpenseReportPage', () => {
  beforeEach(() => {
    resetHarness();
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('exports report as CSV', async () => {
    const now = new Date();
    const selectedMonth = format(now, 'yyyy-MM');
    const start = startOfMonth(new Date(`${selectedMonth}-01`));
    const end = endOfMonth(new Date(`${selectedMonth}-01`));
    setQueryData(['expense-report', start, end], {
      total_amount: 10000,
      total_count: 3,
      category_breakdown: [{ category: 'transportation', amount: 4000 }],
      department_breakdown: [],
      status_summary: {
        draft: 1,
        pending: 1,
        approved: 1,
        rejected: 0,
        reimbursed: 0,
        approved_amount: 5000,
        pending_amount: 3000,
      },
    });
    setQueryData(['expense-monthly-trend', selectedMonth], {
      data: [{ month: '01', amount: 1000 }],
    });
    getApiMocks().expenses.exportCSV.mockResolvedValue(new Blob(['csv']));

    render(<ExpenseReportPage />);

    fireEvent.click(screen.getByRole('button', { name: /CSV/ }));

    await waitFor(() => {
      expect(getApiMocks().expenses.exportCSV).toHaveBeenCalledWith(
        expect.objectContaining({
          start_date: expect.any(String),
          end_date: expect.any(String),
        }),
      );
    });
  });
});
