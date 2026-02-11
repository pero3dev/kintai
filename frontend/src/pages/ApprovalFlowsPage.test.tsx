import { getApiMocks, resetHarness, setQueryData } from './__tests__/testHarness';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { ApprovalFlowsPage } from './ApprovalFlowsPage';

describe('ApprovalFlowsPage', () => {
  beforeEach(() => {
    resetHarness();
  });

  it('renders flows and toggles active state', async () => {
    setQueryData(['approval-flows'], [
      {
        id: 'flow-1',
        name: 'Leave Approval',
        flow_type: 'leave',
        is_active: true,
        steps: [
          {
            id: 'step-1',
            step_order: 1,
            step_type: 'role',
            approver_role: 'manager',
          },
        ],
      },
    ]);

    render(<ApprovalFlowsPage />);

    fireEvent.click(screen.getByRole('button', { name: 'approvalFlows.disable' }));

    await waitFor(() => {
      expect(getApiMocks().approvalFlows.update).toHaveBeenCalledWith('flow-1', {
        is_active: false,
      });
    });
  });
});
