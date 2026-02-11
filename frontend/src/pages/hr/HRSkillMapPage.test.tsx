import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRSkillMapPage } from './HRSkillMapPage';

describe('HRSkillMapPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders employee skill map', () => {
    getApiMocks().hr.getSkillMap.mockReturnValue({
      data: [
        {
          id: 'e1',
          name: 'Alice',
          position: 'Engineer',
          department: 'Dev',
          skills: [{ skill_name: 'Go', category: 'technical', level: 4 }],
        },
      ],
    });
    getApiMocks().hr.getSkillGapAnalysis.mockReturnValue({ data: [] });

    render(<HRSkillMapPage />);

    expect(screen.getByText('Alice')).toBeInTheDocument();
  });
});

