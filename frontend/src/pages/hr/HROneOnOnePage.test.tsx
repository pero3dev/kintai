import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HROneOnOnePage } from './HROneOnOnePage';

describe('HROneOnOnePage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders one-on-one meetings', () => {
    getApiMocks().hr.getOneOnOnes.mockReturnValue({
      data: [
        {
          id: 'm1',
          employee_name: 'Alice',
          scheduled_date: '2026-02-15T10:00:00',
          status: 'scheduled',
          frequency: 'biweekly',
          agenda: 'Career discussion',
          action_items: [],
        },
      ],
    });

    render(<HROneOnOnePage />);

    expect(screen.getByText('Alice')).toBeInTheDocument();
  });
});

