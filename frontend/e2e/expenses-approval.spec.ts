import { test, expect, type Page } from '@playwright/test';

type Role = 'employee' | 'manager';

type User = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: Role;
  is_active: boolean;
};

type ExpenseItem = {
  expense_date: string;
  category: string;
  description: string;
  amount: number;
  receipt_url?: string;
};

type ExpenseRecord = {
  id: string;
  user_id: string;
  user_name: string;
  title: string;
  status: 'draft' | 'pending' | 'approved' | 'rejected';
  notes: string;
  rejected_reason?: string;
  created_at: string;
  amount: number;
  expense_date: string;
  category: string;
  items: ExpenseItem[];
};

type HistoryEntry = {
  action: string;
  details: string;
  user_name: string;
  created_at: string;
};

async function loginAs(page: Page, email: string) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill(email);
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

async function logoutFromLayout(page: Page) {
  const logoutButton = page
    .locator('button:has(span.material-symbols-outlined:has-text("logout")):visible')
    .first();
  await logoutButton.click();
  await expect(page).toHaveURL('/login');
}

async function createExpense(
  page: Page,
  title: string,
  description: string,
  amount: string,
  date: string,
  note: string,
) {
  await page.goto('/expenses/new');
  await expect(page).toHaveURL('/expenses/new');

  await page.locator('input[name="title"]').fill(title);
  await page.locator('input[name="items.0.expense_date"]').fill(date);
  await page.locator('select[name="items.0.category"]').selectOption('meals');
  await page.locator('input[name="items.0.description"]').fill(description);
  await page.locator('input[name="items.0.amount"]').fill(amount);
  await page.locator('textarea[name="notes"]').fill(note);

  await page
    .locator('button[type="submit"]:has(span.material-symbols-outlined:has-text("send"))')
    .click();
  await expect(page).toHaveURL('/expenses');
}

test.describe('Expenses Approval E2E', () => {
  test('should reflect approve/reject results in expense detail and history', async ({ page }) => {
    test.setTimeout(90_000);

    const users: Record<Role, User> = {
      employee: {
        id: 'u-employee',
        email: 'employee@example.com',
        first_name: 'Taro',
        last_name: 'Employee',
        role: 'employee',
        is_active: true,
      },
      manager: {
        id: 'u-manager',
        email: 'manager@example.com',
        first_name: 'Jiro',
        last_name: 'Manager',
        role: 'manager',
        is_active: true,
      },
    };

    const expenses: ExpenseRecord[] = [];
    const histories: Record<string, HistoryEntry[]> = {};
    let currentUser: User = users.employee;
    let approveCalls = 0;
    let rejectCalls = 0;
    let createCalls = 0;

    await page.route('**/api/v1/**', async (route) => {
      const req = route.request();
      const url = req.url();
      const method = req.method();

      if (url.endsWith('/auth/login') && method === 'POST') {
        const body = req.postDataJSON() as { email?: string };
        currentUser = body.email === users.manager.email ? users.manager : users.employee;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: currentUser,
            access_token: `access-${currentUser.role}`,
            refresh_token: `refresh-${currentUser.role}`,
          }),
        });
        return;
      }

      if (url.includes('/attendance/today')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({}),
        });
        return;
      }

      if (url.includes('/notifications?')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (url.includes('/notifications/unread-count')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ count: 0 }),
        });
        return;
      }

      if (url.includes('/expenses/stats')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_this_month: 0,
            pending_count: expenses.filter((e) => e.status === 'pending').length,
            approved_this_month: 0,
            reimbursed_total: 0,
          }),
        });
        return;
      }

      if (url.includes('/expenses/templates')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (url.includes('/expenses/pending') && method === 'GET') {
        const pending = expenses.filter((e) => e.status === 'pending');
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: pending, total: pending.length }),
        });
        return;
      }

      if (url.match(/\/expenses\/[^/]+\/comments$/) && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (url.match(/\/expenses\/[^/]+\/history$/) && method === 'GET') {
        const idMatch = url.match(/\/expenses\/([^/]+)\/history$/);
        const expenseID = idMatch?.[1] || '';
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: histories[expenseID] || [] }),
        });
        return;
      }

      if (url.match(/\/expenses\/[^/]+\/approve$/) && method === 'PUT') {
        const idMatch = url.match(/\/expenses\/([^/]+)\/approve$/);
        const expenseID = idMatch?.[1] || '';
        const body = req.postDataJSON() as { status?: 'approved' | 'rejected'; rejected_reason?: string };
        const target = expenses.find((e) => e.id === expenseID);
        if (target && body.status) {
          target.status = body.status;
          target.rejected_reason = body.rejected_reason;
          histories[expenseID] = histories[expenseID] || [];
          histories[expenseID].push({
            action: body.status === 'approved' ? 'Approved' : 'Rejected',
            details: body.rejected_reason || '',
            user_name: `${currentUser.last_name} ${currentUser.first_name}`,
            created_at: '2026-02-25T10:00:00Z',
          });
        }
        if (body.status === 'approved') {
          approveCalls += 1;
        } else if (body.status === 'rejected') {
          rejectCalls += 1;
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: expenseID, status: body.status }),
        });
        return;
      }

      if (url.match(/\/expenses\/[^/]+$/) && method === 'GET') {
        const idMatch = url.match(/\/expenses\/([^/?]+)$/);
        const expenseID = idMatch?.[1] || '';
        const target = expenses.find((e) => e.id === expenseID);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(target || null),
        });
        return;
      }

      if (url.match(/\/expenses(\?|$)/) && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as {
          title: string;
          status: 'draft' | 'pending';
          notes?: string;
          items: ExpenseItem[];
        };
        const newID = `exp-${expenses.length + 1}`;
        const firstItem = body.items[0];
        const amount = body.items.reduce((sum, i) => sum + (Number(i.amount) || 0), 0);
        const created: ExpenseRecord = {
          id: newID,
          user_id: currentUser.id,
          user_name: `${currentUser.last_name} ${currentUser.first_name}`,
          title: body.title,
          status: body.status,
          notes: body.notes || '',
          created_at: '2026-02-24T09:00:00Z',
          amount,
          expense_date: firstItem?.expense_date || '2026-02-24',
          category: firstItem?.category || 'other',
          items: body.items,
        };
        expenses.push(created);
        histories[newID] = [
          {
            action: 'Submitted',
            details: '',
            user_name: `${currentUser.last_name} ${currentUser.first_name}`,
            created_at: '2026-02-24T09:00:00Z',
          },
        ];

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (url.match(/\/expenses(\?|$)/) && method === 'GET') {
        const mine = expenses.filter((e) => e.user_id === currentUser.id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: mine, total: mine.length }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    const approvedTitle = 'E2E Expense Approve';
    const rejectedTitle = 'E2E Expense Reject';
    const rejectReason = 'Out of policy';

    await loginAs(page, users.employee.email);
    await createExpense(
      page,
      approvedTitle,
      'Team lunch',
      '3000',
      '2026-02-24',
      'expense note 1',
    );
    await createExpense(
      page,
      rejectedTitle,
      'Private dinner',
      '4500',
      '2026-02-24',
      'expense note 2',
    );
    await expect.poll(() => createCalls).toBe(2);

    const approvedID = expenses.find((e) => e.title === approvedTitle)?.id || '';
    const rejectedID = expenses.find((e) => e.title === rejectedTitle)?.id || '';
    expect(approvedID).not.toBe('');
    expect(rejectedID).not.toBe('');

    await logoutFromLayout(page);

    await loginAs(page, users.manager.email);
    await page.goto('/expenses/approve');
    await expect(page).toHaveURL('/expenses/approve');

    const approveRow = page.locator('tr:visible, div.glass-subtle:visible', { hasText: approvedTitle }).first();
    await expect(approveRow).toBeVisible();
    await approveRow.locator('button').first().click();

    const rejectRow = page.locator('tr:visible, div.glass-subtle:visible', { hasText: rejectedTitle }).first();
    await expect(rejectRow).toBeVisible();
    await rejectRow.locator('button').nth(1).click();
    await rejectRow.locator('input').fill(rejectReason);
    await rejectRow.locator('button').first().click();

    await expect.poll(() => approveCalls).toBe(1);
    await expect.poll(() => rejectCalls).toBe(1);

    await logoutFromLayout(page);

    await loginAs(page, users.employee.email);

    await page.goto(`/expenses/${approvedID}`);
    await expect(page).toHaveURL(`/expenses/${approvedID}`);
    await expect(page.locator('span[class*="bg-green-500/20"]').first()).toBeVisible();
    await page.locator('button:has(span.material-symbols-outlined:has-text("history"))').click();
    await expect(page.locator('*:visible', { hasText: 'Approved' }).first()).toBeVisible();

    await page.goto(`/expenses/${rejectedID}`);
    await expect(page).toHaveURL(`/expenses/${rejectedID}`);
    await expect(page.locator('span[class*="bg-red-500/20"]').first()).toBeVisible();
    await expect(page.locator('*:visible', { hasText: rejectReason }).first()).toBeVisible();
    await page.locator('button:has(span.material-symbols-outlined:has-text("history"))').click();
    await expect(page.locator('*:visible', { hasText: 'Rejected' }).first()).toBeVisible();
  });
});
