import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRGoalsPage } from './HRGoalsPage';

describe('HRGoalsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders goals list', () => {
    getApiMocks().hr.getGoals.mockReturnValue({
      data: [
        {
          id: 'g1',
          title: 'Improve quality',
          employee_name: 'Alice',
          category: 'performance',
          priority: 'high',
          status: 'in_progress',
          progress: 40,
        },
      ],
    });

    render(<HRGoalsPage />);

    expect(screen.getByText('Improve quality')).toBeInTheDocument();
  });
});

