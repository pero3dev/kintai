import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HREvaluationsPage } from './HREvaluationsPage';

describe('HREvaluationsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders evaluation cards', () => {
    getApiMocks().hr.getEvaluationCycles.mockReturnValue({
      data: [{ id: 'c1', name: '2026 H1', status: 'active' }],
    });
    getApiMocks().hr.getEvaluations.mockReturnValue({
      data: [
        {
          id: 'ev-1',
          employee_name: 'Alice',
          cycle_name: '2026 H1',
          status: 'draft',
          final_score: 'A',
        },
      ],
    });

    render(<HREvaluationsPage />);

    expect(screen.getByText('Alice')).toBeInTheDocument();
  });
});

