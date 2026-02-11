import { resetHarness, setQueryData } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseHistoryPage } from './ExpenseHistoryPage';

describe('ExpenseHistoryPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders expense history list', () => {
    setQueryData(['expenses', 'history', 1, 10, '', ''], {
      data: [
        {
          id: 'exp-1',
          title: 'Taxi',
          expense_date: '2026-02-10',
          category: 'transportation',
          amount: 1800,
          status: 'pending',
        },
      ],
      total: 1,
    });

    render(<ExpenseHistoryPage />);

    expect(screen.getByText('expenses.history.title')).toBeInTheDocument();
    expect(screen.getAllByText('Taxi').length).toBeGreaterThan(0);
  });
});
