import {
  getApiMocks,
  resetHarness,
  setUserRole,
} from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ExportPage } from './ExportPage';

describe('ExportPage', () => {
  beforeEach(() => {
    resetHarness();
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('runs attendance CSV export for admin role', async () => {
    setUserRole('admin');
    getApiMocks().export.attendance.mockResolvedValue(new Blob(['csv']));

    render(<ExportPage />);

    fireEvent.click(screen.getAllByRole('button', { name: 'CSV' })[0]);

    await waitFor(() => {
      expect(getApiMocks().export.attendance).toHaveBeenCalledWith(
        expect.objectContaining({
          start_date: expect.any(String),
          end_date: expect.any(String),
        }),
      );
    });
  });
});
