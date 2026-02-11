import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRAttendanceIntegrationPage } from './HRAttendanceIntegrationPage';

describe('HRAttendanceIntegrationPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders attendance integration data', () => {
    getApiMocks().hr.getAttendanceIntegration.mockReturnValue({
      data: {
        avg_work_hours: 7.8,
        total_overtime: 120,
        late_rate: 4.1,
        leave_usage: 56,
        employees: [
          {
            id: 'e1',
            name: 'Alice',
            department: 'Dev',
            overtime_hours: 12,
            late_count: 1,
            leave_usage: 34,
            absent_days: 0,
            risk_level: 'low',
          },
        ],
      },
    });
    getApiMocks().hr.getAttendanceAlerts.mockReturnValue({
      data: [{ type: 'late', severity: 'medium', employee_name: 'Alice', message: 'Late arrivals' }],
    });
    getApiMocks().hr.getAttendanceTrend.mockReturnValue({
      data: [{ label: 'W1', overtime_hours: 10 }],
    });

    render(<HRAttendanceIntegrationPage />);

    expect(screen.getAllByText('Alice').length).toBeGreaterThan(0);
  });
});

