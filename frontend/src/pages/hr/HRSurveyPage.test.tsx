import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRSurveyPage } from './HRSurveyPage';

describe('HRSurveyPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders surveys list', () => {
    getApiMocks().hr.getSurveys.mockReturnValue({
      data: [
        {
          id: 's1',
          title: 'Engagement Survey',
          type: 'engagement',
          status: 'draft',
          questions: [{ id: 'q1', text: 'How are you?' }],
          response_count: 0,
        },
      ],
    });

    render(<HRSurveyPage />);

    expect(screen.getByText('Engagement Survey')).toBeInTheDocument();
  });
});

