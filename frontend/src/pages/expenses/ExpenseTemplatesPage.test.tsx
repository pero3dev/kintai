import { getApiMocks, resetHarness, setQueryData } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpenseTemplatesPage } from './ExpenseTemplatesPage';

describe('ExpenseTemplatesPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('uses a saved template', async () => {
    setQueryData(['expense-templates'], {
      data: [
        {
          id: 'tpl-1',
          name: 'Taxi template',
          title: 'Taxi',
          category: 'transportation',
          amount: 1200,
          is_recurring: false,
        },
      ],
    });

    render(<ExpenseTemplatesPage />);

    fireEvent.click(screen.getByRole('button', { name: /expenses\.templates\.useTemplate/ }));

    await waitFor(() => {
      expect(getApiMocks().expenses.useTemplate).toHaveBeenCalledWith('tpl-1');
    });
  });
});
