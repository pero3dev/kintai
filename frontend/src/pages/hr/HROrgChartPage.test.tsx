import { getApiMocks, resetHarness } from '../__tests__/testHarness';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { HROrgChartPage } from './HROrgChartPage';

describe('HROrgChartPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders org chart tree', () => {
    getApiMocks().hr.getOrgChart.mockReturnValue({
      data: {
        id: 'root',
        name: 'CEO',
        position: 'Chief Executive Officer',
        department: 'Executive',
        children: [],
      },
    });

    render(<HROrgChartPage />);

    expect(screen.getByText('CEO')).toBeInTheDocument();
  });
});

