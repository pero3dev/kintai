import { test, expect, type Page } from '@playwright/test';

type Policy = {
  id: string;
  category: string;
  monthly_limit: number;
  per_claim_limit: number;
  auto_approve_limit: number;
  requires_receipt_above: number;
  description: string;
  is_active: boolean;
};

async function loginAsAdmin(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('admin@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Expense Policy E2E', () => {
  test('should create, update, and delete policy as admin', async ({ page }) => {
    test.setTimeout(90_000);

    let policies: Policy[] = [
      {
        id: 'pol-1',
        category: 'transportation',
        monthly_limit: 50000,
        per_claim_limit: 5000,
        auto_approve_limit: 3000,
        requires_receipt_above: 1000,
        description: 'Default transportation policy',
        is_active: true,
      },
    ];

    const budgets = [
      { department: 'Engineering', used_amount: 250000, budget_amount: 500000 },
      { department: 'HR', used_amount: 120000, budget_amount: 200000 },
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
              id: 'u-admin',
              email: 'admin@example.com',
              first_name: 'Admin',
              last_name: 'User',
              role: 'admin',
              is_active: true,
            },
            access_token: 'access-admin',
            refresh_token: 'refresh-admin',
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

      if (path.endsWith('/api/v1/notifications') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (path.endsWith('/api/v1/notifications/unread-count') && method === 'GET') {
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

      if (path.endsWith('/expenses/policies') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: policies }),
        });
        return;
      }

      if (path.endsWith('/expenses/policies') && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const created: Policy = {
          id: `pol-${policies.length + 1}`,
          category: String(body.category || ''),
          monthly_limit: Number(body.monthly_limit || 0),
          per_claim_limit: Number(body.per_claim_limit || 0),
          auto_approve_limit: Number(body.auto_approve_limit || 0),
          requires_receipt_above: Number(body.requires_receipt_above || 0),
          description: String(body.description || ''),
          is_active: Boolean(body.is_active),
        };
        policies = [...policies, created];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (/\/expenses\/policies\/[^/]+$/.test(path) && method === 'PUT') {
        updateCalls += 1;
        const id = path.split('/').pop() || '';
        const body = req.postDataJSON() as Record<string, unknown>;
        policies = policies.map((p) =>
          p.id === id
            ? {
                ...p,
                category: String(body.category ?? p.category),
                monthly_limit: Number(body.monthly_limit ?? p.monthly_limit),
                per_claim_limit: Number(body.per_claim_limit ?? p.per_claim_limit),
                auto_approve_limit: Number(body.auto_approve_limit ?? p.auto_approve_limit),
                requires_receipt_above: Number(body.requires_receipt_above ?? p.requires_receipt_above),
                description: String(body.description ?? p.description),
                is_active: Boolean(body.is_active ?? p.is_active),
              }
            : p,
        );
        const updated = policies.find((p) => p.id === id) || null;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(updated),
        });
        return;
      }

      if (/\/expenses\/policies\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const id = path.split('/').pop() || '';
        policies = policies.filter((p) => p.id !== id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (path.endsWith('/expenses/budgets') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: budgets }),
        });
        return;
      }

      if (path.endsWith('/expenses/policy-violations') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    await loginAsAdmin(page);
    await page.goto('/expenses/policy');
    await expect(page).toHaveURL('/expenses/policy');

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("add")):visible')
      .first()
      .click();

    const form = page.locator('div.glass-card.rounded-2xl.p-6.space-y-4').first();
    await expect(form).toBeVisible();

    await form.locator('select').first().selectOption('meals');
    const numberInputs = form.locator('input[type="number"]');
    await numberInputs.nth(0).fill('80000');
    await numberInputs.nth(1).fill('8000');
    await numberInputs.nth(2).fill('5000');
    await numberInputs.nth(3).fill('2000');
    await form
      .locator('input:not([type="number"]):not([type="checkbox"])')
      .first()
      .fill('E2E meals policy');
    await form.locator('button.gradient-primary').click();
    await expect.poll(() => createCalls).toBe(1);

    let policyCard = page
      .locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E meals policy' })
      .first();
    await expect(policyCard).toBeVisible();

    await policyCard
      .locator('button:has(span.material-symbols-outlined:has-text("edit"))')
      .first()
      .click();

    const editForm = page.locator('div.glass-card.rounded-2xl.p-6.space-y-4').first();
    await editForm
      .locator('input:not([type="number"]):not([type="checkbox"])')
      .first()
      .fill('E2E meals policy updated');
    await editForm.locator('button.gradient-primary').click();
    await expect.poll(() => updateCalls).toBe(1);

    policyCard = page
      .locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E meals policy updated' })
      .first();
    await expect(policyCard).toBeVisible();

    await policyCard
      .locator('button:has(span.material-symbols-outlined:has-text("delete"))')
      .first()
      .click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page.locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E meals policy updated' })).toHaveCount(0);
  });
});
