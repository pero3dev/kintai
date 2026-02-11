import { getApiMocks, resetHarness, setQueryData, setUserRole } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ExpensePolicyPage } from './ExpensePolicyPage';

describe('ExpensePolicyPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('creates a policy as admin', async () => {
    setUserRole('admin');
    setQueryData(['expense-policies'], { data: [] });
    setQueryData(['expense-budgets'], { data: [] });
    setQueryData(['expense-policy-violations'], { data: [] });

    const { container } = render(<ExpensePolicyPage />);

    fireEvent.click(screen.getByRole('button', { name: /expenses\.policy\.addPolicy/ }));

    const select = container.querySelector('select');
    if (!select) throw new Error('Category select was not found');
    fireEvent.change(select, { target: { value: 'transportation' } });
    fireEvent.click(screen.getByRole('button', { name: /common\.save/ }));

    await waitFor(() => {
      expect(getApiMocks().expenses.createPolicy).toHaveBeenCalledWith(
        expect.objectContaining({ category: 'transportation' }),
      );
    });
  });
});
