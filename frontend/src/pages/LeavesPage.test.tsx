import { resetHarness, setQueryData, setUserRole } from './__tests__/testHarness';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { LeavesPage } from './LeavesPage';

describe('LeavesPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('opens request form and shows validation errors', async () => {
    setUserRole('employee');
    setQueryData(['leaves', 'my', 1, 20], { data: [], total_pages: 0, total: 0 });

    render(<LeavesPage />);

    fireEvent.click(screen.getByRole('button', { name: 'leaves.newRequest' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));

    expect(await screen.findByText('leaves.validation.startDateRequired')).toBeInTheDocument();
  });
});
