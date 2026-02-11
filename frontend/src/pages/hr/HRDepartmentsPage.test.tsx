import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRDepartmentsPage } from './HRDepartmentsPage';

describe('HRDepartmentsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders department cards', () => {
    getApiMocks().hr.getDepartments.mockReturnValue({
      data: [
        {
          id: 'd1',
          name: 'Engineering',
          code: 'ENG',
          description: 'Core development',
          manager_name: 'Alice',
          member_count: 12,
        },
      ],
    });

    render(<HRDepartmentsPage />);

    expect(screen.getByText('Engineering')).toBeInTheDocument();
  });
});

