import { test, expect, type Page } from '@playwright/test';

type FlowStep = {
  id: string;
  step_order: number;
  step_type: 'role' | 'specific_user';
  approver_role?: string;
  approver_id?: string;
};

type ApprovalFlow = {
  id: string;
  name: string;
  flow_type: 'leave' | 'overtime' | 'correction';
  is_active: boolean;
  steps: FlowStep[];
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Approval Flows E2E', () => {
  test('should create, toggle, and delete approval flow', async ({ page }) => {
    test.setTimeout(90_000);

    let flows: ApprovalFlow[] = [
      {
        id: 'flow-1',
        name: 'Default Leave Flow',
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
    ];

    let createCalls = 0;
    let updateCalls = 0;
    let deleteCalls = 0;

    await page.route('**/api/v1/**', async (route) => {
      const req = route.request();
      const url = req.url();
      const method = req.method();
      const path = new URL(url).pathname;

      if (path.endsWith('/auth/login') && method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: {
              id: 'u-manager',
              email: 'manager@example.com',
              first_name: 'Jiro',
              last_name: 'Manager',
              role: 'manager',
              is_active: true,
            },
            access_token: 'access-manager',
            refresh_token: 'refresh-manager',
          }),
        });
        return;
      }

      if (path.endsWith('/attendance/today') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({}),
        });
        return;
      }

      if (path.endsWith('/notifications') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (path.endsWith('/notifications/unread-count') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ count: 0 }),
        });
        return;
      }

      if (path.endsWith('/leaves/pending') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (path.endsWith('/approval-flows') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(flows),
        });
        return;
      }

      if (path.endsWith('/approval-flows') && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const steps = Array.isArray(body.steps) ? (body.steps as Array<Record<string, unknown>>) : [];
        const created: ApprovalFlow = {
          id: `flow-${flows.length + 1}`,
          name: String(body.name || ''),
          flow_type: (body.flow_type as ApprovalFlow['flow_type']) || 'leave',
          is_active: true,
          steps: steps.map((s, index) => ({
            id: `step-new-${index + 1}`,
            step_order: Number(s.step_order || index + 1),
            step_type: (s.step_type as FlowStep['step_type']) || 'role',
            approver_role: String(s.approver_role || ''),
            approver_id: String(s.approver_id || ''),
          })),
        };
        flows = [...flows, created];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (/\/approval-flows\/[^/]+$/.test(path) && method === 'PUT') {
        updateCalls += 1;
        const id = path.split('/').pop() || '';
        const body = req.postDataJSON() as Record<string, unknown>;
        flows = flows.map((f) =>
          f.id === id
            ? {
                ...f,
                is_active: typeof body.is_active === 'boolean' ? body.is_active : f.is_active,
              }
            : f,
        );
        const updated = flows.find((f) => f.id === id) || null;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(updated),
        });
        return;
      }

      if (/\/approval-flows\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const id = path.split('/').pop() || '';
        flows = flows.filter((f) => f.id !== id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    await loginAsManager(page);
    await page.goto('/approval-flows');
    await expect(page).toHaveURL('/approval-flows');

    await page.locator('button:has(svg.lucide-plus):visible').first().click();
    const form = page.locator('form').first();
    await expect(form).toBeVisible();

    await form.locator('input[name="name"]').fill('E2E Overtime Flow');
    await form.locator('select[name="flow_type"]').selectOption('overtime');
    await form.locator('button[type="submit"]').click();
    await expect.poll(() => createCalls).toBe(1);

    let flowCard = page.locator('div.glass-card.rounded-2xl.p-6', { hasText: 'E2E Overtime Flow' }).first();
    await expect(flowCard).toBeVisible();

    const actionButtons = flowCard.locator('div.flex.gap-2 button');
    await actionButtons.first().click();
    await expect.poll(() => updateCalls).toBe(1);

    flowCard = page.locator('div.glass-card.rounded-2xl.p-6', { hasText: 'E2E Overtime Flow' }).first();
    await flowCard.locator('div.flex.gap-2 button').nth(1).click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page.locator('div.glass-card.rounded-2xl.p-6', { hasText: 'E2E Overtime Flow' })).toHaveCount(0);
  });
});
