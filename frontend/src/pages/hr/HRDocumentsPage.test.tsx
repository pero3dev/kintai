import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HRDocumentsPage } from './HRDocumentsPage';

describe('HRDocumentsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders uploaded documents', () => {
    getApiMocks().hr.getDocuments.mockReturnValue({
      data: [
        {
          id: 'doc-1',
          title: 'Contract.pdf',
          type: 'contract',
          file_size: 1024,
          uploaded_at: '2026-02-10',
        },
      ],
    });

    render(<HRDocumentsPage />);

    expect(screen.getAllByText('Contract.pdf').length).toBeGreaterThan(0);
  });
});

