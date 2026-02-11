import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HROnboardingPage } from './HROnboardingPage';

describe('HROnboardingPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders onboarding list', () => {
    getApiMocks().hr.getOnboardingTemplates.mockReturnValue({
      data: [{ id: 'tpl-1', name: 'Default' }],
    });
    getApiMocks().hr.getOnboardings.mockReturnValue({
      data: [
        {
          id: 'onb1',
          employee_name: 'Alice',
          start_date: '2026-02-10',
          status: 'in_progress',
          tasks: [{ id: 't1', title: 'Prepare PC', completed: false }],
        },
      ],
    });

    render(<HROnboardingPage />);

    expect(screen.getByText('Alice')).toBeInTheDocument();
  });
});

