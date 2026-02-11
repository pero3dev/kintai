import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRAnnouncementsPage } from './HRAnnouncementsPage';

describe('HRAnnouncementsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders announcement cards', () => {
    getApiMocks().hr.getAnnouncements.mockReturnValue({
      data: [
        {
          id: 'a1',
          title: 'Office Update',
          content: 'Please check the new policy.',
          priority: 'high',
          target: 'all',
          author_name: 'HR Team',
        },
      ],
    });

    render(<HRAnnouncementsPage />);

    expect(screen.getByText('Office Update')).toBeInTheDocument();
  });
});

