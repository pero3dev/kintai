import { test, expect, type Page } from '@playwright/test';

type ExpenseRecord = {
  id: string;
  title: string;
  user_name: string;
  amount: number;
  status: 'pending' | 'approved' | 'returned' | 'rejected';
  current_step: number;
  created_at: string;
};

type DelegateRecord = {
  id: string;
  delegate_to: string;
  delegate_name: string;
  start_date: string;
  end_date: string;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Expenses Advanced Approve E2E', () => {
  test('should approve/return/reject pending expenses and manage delegates', async ({ page }) => {
    test.setTimeout(90_000);

    let expenses: ExpenseRecord[] = [
      {
        id: 'exp-1',
        title: 'E2E Advanced Approve',
        user_name: 'Yamada Taro',
        amount: 12000,
        status: 'pending',
        current_step: 1,
        created_at: '2026-03-01T09:00:00Z',
      },
      {
        id: 'exp-2',
        title: 'E2E Advanced Return',
        user_name: 'Suzuki Hanako',
        amount: 28000,
        status: 'pending',
        current_step: 2,
        created_at: '2026-03-02T09:00:00Z',
      },
      {
        id: 'exp-3',
        title: 'E2E Advanced Reject',
        user_name: 'Sato Ken',
        amount: 45000,
        status: 'pending',
        current_step: 2,
        created_at: '2026-03-03T09:00:00Z',
      },
    ];

    let delegates: DelegateRecord[] = [
      {
        id: 'del-1',
        delegate_to: 'u-admin',
        delegate_name: 'Admin User',
        start_date: '2026-03-01',
        end_date: '2026-03-05',
      },
    ];

    let approveCalls = 0;
    let returnCalls = 0;
    let rejectCalls = 0;
    let setDelegateCalls = 0;
    let removeDelegateCalls = 0;

    await page.route('**/api/v1/**', async (route) => {
      const req = route.request();
      const method = req.method();
      const url = req.url();
      const parsed = new URL(url);
      const path = parsed.pathname;

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

      if (path.endsWith('/expenses/pending') && method === 'GET') {
        const pendingExpenses = expenses.filter((e) => e.status === 'pending');
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: pendingExpenses,
            total: pendingExpenses.length,
          }),
        });
        return;
      }

      if (path.endsWith('/expenses/approval-flow') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            steps: [
              { name: 'Manager', approver_role: 'manager', auto_approve_below: 0 },
              { name: 'Admin', approver_role: 'admin', auto_approve_below: 10000 },
            ],
          }),
        });
        return;
      }

      if (path.endsWith('/expenses/delegates') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: delegates }),
        });
        return;
      }

      if (path.endsWith('/users') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [
              {
                id: 'u-admin',
                first_name: 'Admin',
                last_name: 'User',
                role: 'admin',
              },
              {
                id: 'u-manager-2',
                first_name: 'Backup',
                last_name: 'Manager',
                role: 'manager',
              },
              {
                id: 'u-employee',
                first_name: 'General',
                last_name: 'Employee',
                role: 'employee',
              },
            ],
          }),
        });
        return;
      }

      if (/\/expenses\/[^/]+\/advanced-approve$/.test(path) && method === 'PUT') {
        const id = path.split('/').slice(-2)[0] || '';
        const body = req.postDataJSON() as { action?: 'approve' | 'return' | 'reject'; reason?: string };

        if (body.action === 'approve') {
          approveCalls += 1;
          expenses = expenses.map((e) => (e.id === id ? { ...e, status: 'approved' } : e));
        }
        if (body.action === 'return') {
          returnCalls += 1;
          expenses = expenses.map((e) => (e.id === id ? { ...e, status: 'returned' } : e));
        }
        if (body.action === 'reject') {
          rejectCalls += 1;
          expenses = expenses.map((e) => (e.id === id ? { ...e, status: 'rejected' } : e));
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id, action: body.action, reason: body.reason || '' }),
        });
        return;
      }

      if (path.endsWith('/expenses/delegates') && method === 'POST') {
        setDelegateCalls += 1;
        const body = req.postDataJSON() as { delegate_to?: string; start_date?: string; end_date?: string };
        const candidateName = body.delegate_to === 'u-admin' ? 'Admin User' : 'Backup Manager';
        delegates = [
          ...delegates,
          {
            id: `del-${delegates.length + 1}`,
            delegate_to: body.delegate_to || '',
            delegate_name: candidateName,
            start_date: body.start_date || '',
            end_date: body.end_date || '',
          },
        ];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (/\/expenses\/delegates\/[^/]+$/.test(path) && method === 'DELETE') {
        removeDelegateCalls += 1;
        const id = path.split('/').pop() || '';
        delegates = delegates.filter((d) => d.id !== id);
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
    await page.goto('/expenses/advanced-approve');
    await expect(page).toHaveURL('/expenses/advanced-approve');

    await expect(page.locator('tr', { hasText: 'E2E Advanced Approve' }).first()).toBeVisible();
    await expect(page.locator('tr', { hasText: 'E2E Advanced Return' }).first()).toBeVisible();
    await expect(page.locator('tr', { hasText: 'E2E Advanced Reject' }).first()).toBeVisible();

    const approveRow = page.locator('tr', { hasText: 'E2E Advanced Approve' }).first();
    await approveRow.locator('button').nth(0).click();
    await expect.poll(() => approveCalls).toBe(1);
    await expect(page.locator('tr', { hasText: 'E2E Advanced Approve' })).toHaveCount(0);

    const returnRow = page.locator('tr', { hasText: 'E2E Advanced Return' }).first();
    await returnRow.locator('button').nth(1).click();
    const returnReason = 'Need more receipt details';
    await returnRow.locator('textarea').fill(returnReason);
    await returnRow.locator('button').nth(0).click();
    await expect.poll(() => returnCalls).toBe(1);
    await expect(page.locator('tr', { hasText: 'E2E Advanced Return' })).toHaveCount(0);

    const rejectRow = page.locator('tr', { hasText: 'E2E Advanced Reject' }).first();
    await rejectRow.locator('button').nth(2).click();
    const rejectReason = 'Policy violation';
    await rejectRow.locator('textarea').fill(rejectReason);
    await rejectRow.locator('button').nth(0).click();
    await expect.poll(() => rejectCalls).toBe(1);
    await expect(page.locator('tr', { hasText: 'E2E Advanced Reject' })).toHaveCount(0);

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("swap_horiz"))')
      .first()
      .click();

    const delegateSelect = page.locator('select').first();
    await delegateSelect.selectOption('u-manager-2');
    const dateInputs = page.locator('input[type="date"]');
    await dateInputs.nth(0).fill('2026-03-10');
    await dateInputs.nth(1).fill('2026-03-20');
    await page.locator('button.gradient-primary').first().click();
    await expect.poll(() => setDelegateCalls).toBe(1);

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("swap_horiz"))')
      .first()
      .click();
    await expect(page.locator('*:visible', { hasText: 'Backup Manager' }).first()).toBeVisible();

    const newDelegateRow = page.locator('div.glass-subtle.rounded-xl.p-3', { hasText: 'Backup Manager' }).first();
    await newDelegateRow.locator('button:has(span.material-symbols-outlined:has-text("delete"))').click();
    await expect.poll(() => removeDelegateCalls).toBe(1);
    await expect(page.locator('*:visible', { hasText: 'Backup Manager' })).toHaveCount(0);
  });
});
