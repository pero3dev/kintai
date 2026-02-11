import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HREmployeesPage } from './HREmployeesPage';

describe('HREmployeesPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders employees list', () => {
    getApiMocks().hr.getDepartments.mockReturnValue({
      data: [{ id: 'd1', name: 'Dev' }],
    });
    getApiMocks().hr.getEmployees.mockReturnValue({
      data: [
        {
          id: 'e1',
          employee_id: 'EMP-001',
          first_name: 'Taro',
          last_name: 'Yamada',
          department_name: 'Dev',
          position: 'Engineer',
          employment_type: 'fullTime',
          status: 'active',
          hire_date: '2026-01-01',
        },
      ],
      total: 1,
    });

    render(<HREmployeesPage />);

    expect(screen.getAllByText('Yamada Taro').length).toBeGreaterThan(0);
  });
});

