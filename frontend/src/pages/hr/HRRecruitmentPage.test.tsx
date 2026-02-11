import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRRecruitmentPage } from './HRRecruitmentPage';

describe('HRRecruitmentPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders open positions', () => {
    getApiMocks().hr.getPositions.mockReturnValue({
      data: [
        {
          id: 'p1',
          title: 'Frontend Engineer',
          department: 'Dev',
          location: 'Tokyo',
          status: 'open',
          employment_type: 'fullTime',
        },
      ],
    });
    getApiMocks().hr.getApplicants.mockReturnValue({ data: [] });

    render(<HRRecruitmentPage />);

    expect(screen.getByText('Frontend Engineer')).toBeInTheDocument();
  });
});

