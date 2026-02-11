import {
  getApiMocks,
  resetHarness,
  setQueryData,
  setUserRole,
} from './__tests__/testHarness';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ProjectsPage } from './ProjectsPage';

describe('ProjectsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders projects and submits time-entry form', async () => {
    setUserRole('manager');
    setQueryData(['projects', 1, 12], {
      data: [
        {
          id: 'p-1',
          code: 'PRJ-001',
          name: 'Core Platform',
          status: 'active',
          budget_hours: 100,
        },
      ],
      total_pages: 1,
      total: 1,
    });
    setQueryData(['timeEntries', 'my'], { data: [], total_pages: 0, total: 0 });
    setQueryData(['timeEntries', 'summary'], { data: [] });

    render(<ProjectsPage />);

    expect(screen.getByText('Core Platform')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'projects.logTime' }));
    fireEvent.change(document.querySelector('select[name=\"project_id\"]') as HTMLSelectElement, {
      target: { value: 'p-1' },
    });
    fireEvent.change(document.querySelector('input[name=\"date\"]') as HTMLInputElement, {
      target: { value: '2026-02-11' },
    });
    fireEvent.change(document.querySelector('input[name=\"minutes\"]') as HTMLInputElement, {
      target: { value: '60' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));

    expect(await screen.findByText('projects.title')).toBeInTheDocument();
    expect(getApiMocks().timeEntries.create).toHaveBeenCalledWith(
      expect.objectContaining({
        project_id: 'p-1',
        date: '2026-02-11',
        minutes: 60,
      }),
    );
  });
});
