import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRTrainingPage } from './HRTrainingPage';

describe('HRTrainingPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders training programs', () => {
    getApiMocks().hr.getTrainingPrograms.mockReturnValue({
      data: [
        {
          id: 'tr1',
          title: 'Go Basics',
          category: 'technical',
          status: 'upcoming',
          instructor: 'Alice',
          start_date: '2026-02-10',
          end_date: '2026-02-12',
          enrolled_count: 3,
          max_participants: 10,
          is_enrolled: false,
        },
      ],
    });

    render(<HRTrainingPage />);

    expect(screen.getByText('Go Basics')).toBeInTheDocument();
  });
});

