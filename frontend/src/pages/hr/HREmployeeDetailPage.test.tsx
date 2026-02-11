import { getApiMocks, resetHarness, setRouteParams } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HREmployeeDetailPage } from './HREmployeeDetailPage';

describe('HREmployeeDetailPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders employee profile', () => {
    setRouteParams({ employeeId: 'emp-1' });
    getApiMocks().hr.getEmployee.mockReturnValue({
      data: {
        id: 'emp-1',
        employee_id: 'EMP-001',
        first_name: 'Taro',
        last_name: 'Yamada',
        email: 'taro@example.com',
        department_name: 'Dev',
        position: 'Engineer',
        skills: ['Go'],
      },
    });

    render(<HREmployeeDetailPage />);

    expect(screen.getAllByText('Yamada Taro').length).toBeGreaterThan(0);
  });
});

