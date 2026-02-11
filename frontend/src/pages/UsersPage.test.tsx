import { resetHarness, setQueryData } from './__tests__/testHarness';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { UsersPage } from './UsersPage';

describe('UsersPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('opens create modal and validates required fields', async () => {
    setQueryData(['users', 1, 20], {
      data: [
        {
          id: 'u-1',
          email: 'user@example.com',
          first_name: 'Taro',
          last_name: 'Yamada',
          role: 'employee',
          is_active: true,
        },
      ],
      total_pages: 1,
      total: 1,
    });
    setQueryData(['departments'], []);

    render(<UsersPage />);

    fireEvent.click(screen.getByRole('button', { name: 'users.addNew' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));

    expect(await screen.findByText('users.requiredFieldsError')).toBeInTheDocument();
  });
});
