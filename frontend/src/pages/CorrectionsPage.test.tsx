import {
  getApiMocks,
  resetHarness,
  setQueryData,
  setUserRole,
} from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { CorrectionsPage } from './CorrectionsPage';

describe('CorrectionsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders manager pending list and approves request', async () => {
    setUserRole('manager');
    setQueryData(['corrections', 'my', 1, 20], { data: [], total_pages: 0, total: 0 });
    setQueryData(['corrections', 'pending', 1], {
      data: [
        {
          id: 'c-1',
          user: { last_name: 'Suzuki', first_name: 'Hanako' },
          date: '2026-02-11',
          reason: 'Fix missed punch',
        },
      ],
      total_pages: 1,
      total: 1,
    });

    render(<CorrectionsPage />);

    fireEvent.click(screen.getAllByText('common.approve')[0]);

    await waitFor(() => {
      expect(getApiMocks().corrections.approve).toHaveBeenCalledWith('c-1', { status: 'approved' });
    });
  });
});
