import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HROffboardingPage } from './HROffboardingPage';

describe('HROffboardingPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders offboarding list', () => {
    getApiMocks().hr.getOffboardings.mockReturnValue({
      data: [
        {
          id: 'ob1',
          employee_name: 'Alice',
          reason: 'resignation',
          last_working_date: '2026-02-20',
          status: 'pending',
          checklist: [{ key: 'asset_return', completed: false }],
        },
      ],
    });

    render(<HROffboardingPage />);

    expect(screen.getByText('Alice')).toBeInTheDocument();
  });
});

