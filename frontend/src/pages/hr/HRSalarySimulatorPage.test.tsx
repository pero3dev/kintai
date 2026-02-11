import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRSalarySimulatorPage } from './HRSalarySimulatorPage';

describe('HRSalarySimulatorPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('runs salary simulation', async () => {
    getApiMocks().hr.getSalaryOverview.mockReturnValue({
      data: {
        avg_salary: 4000000,
        median_salary: 3800000,
        total_payroll: 80000000,
        headcount: 20,
        department_breakdown: [{ name: 'Dev', avg_salary: 4500000, headcount: 10 }],
      },
    });
    getApiMocks().hr.getBudgetOverview.mockReturnValue({
      data: {
        total_budget: 100000000,
        used_budget: 70000000,
        departments: [{ name: 'Dev', usage_rate: 70 }],
      },
    });

    render(<HRSalarySimulatorPage />);

    fireEvent.change(screen.getByPlaceholderText('hr.salary.grade'), { target: { value: 'G5' } });
    fireEvent.change(screen.getByPlaceholderText('hr.salary.position'), { target: { value: 'Senior' } });
    fireEvent.change(screen.getByPlaceholderText('hr.salary.evaluationScore'), { target: { value: '4.5' } });
    fireEvent.change(screen.getByPlaceholderText('hr.salary.yearsOfService'), { target: { value: '6' } });
    fireEvent.click(screen.getByRole('button', { name: /hr\.salary\.runSimulation/ }));

    await waitFor(() => {
      expect(getApiMocks().hr.simulateSalary).toHaveBeenCalled();
    });
  });
});

