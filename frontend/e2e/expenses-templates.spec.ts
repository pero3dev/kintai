import { test, expect, type Page } from '@playwright/test';

type Template = {
  id: string;
  name: string;
  title: string;
  category: string;
  description: string;
  amount: number;
  is_recurring: boolean;
  recurring_day: number;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Expense Templates E2E', () => {
  test('should create, use, update, and delete template', async ({ page }) => {
    test.setTimeout(90_000);

    let templates: Template[] = [
      {
        id: 'tpl-1',
        name: 'Taxi Basic',
        title: 'Taxi fare',
        category: 'transportation',
        description: 'Default taxi fare',
        amount: 1200,
        is_recurring: false,
        recurring_day: 1,
      },
    ];

    let createCalls = 0;
    let useCalls = 0;
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

      if (path.endsWith('/expenses/templates') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: templates }),
        });
        return;
      }

      if (path.endsWith('/expenses/templates') && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const created: Template = {
          id: `tpl-${templates.length + 1}`,
          name: String(body.name || ''),
          title: String(body.title || ''),
          category: String(body.category || ''),
          description: String(body.description || ''),
          amount: Number(body.amount || 0),
          is_recurring: Boolean(body.is_recurring),
          recurring_day: Number(body.recurring_day || 1),
        };
        templates = [...templates, created];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (/\/expenses\/templates\/[^/]+\/use$/.test(path) && method === 'POST') {
        useCalls += 1;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({}),
        });
        return;
      }

      if (/\/expenses\/templates\/[^/]+$/.test(path) && method === 'PUT') {
        updateCalls += 1;
        const id = path.split('/').pop() || '';
        const body = req.postDataJSON() as Record<string, unknown>;
        templates = templates.map((tpl) =>
          tpl.id === id
            ? {
                ...tpl,
                name: String(body.name ?? tpl.name),
                title: String(body.title ?? tpl.title),
                category: String(body.category ?? tpl.category),
                description: String(body.description ?? tpl.description),
                amount: Number(body.amount ?? tpl.amount),
                is_recurring: Boolean(body.is_recurring ?? tpl.is_recurring),
                recurring_day: Number(body.recurring_day ?? tpl.recurring_day),
              }
            : tpl,
        );
        const updated = templates.find((tpl) => tpl.id === id) || null;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(updated),
        });
        return;
      }

      if (/\/expenses\/templates\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const id = path.split('/').pop() || '';
        templates = templates.filter((tpl) => tpl.id !== id);
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

    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    await loginAsManager(page);
    await page.goto('/expenses/templates');
    await expect(page).toHaveURL('/expenses/templates');

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("add")):visible')
      .first()
      .click();

    const form = page.locator('div.glass-card.rounded-2xl.p-6.space-y-4').first();
    await expect(form).toBeVisible();
    const textInputs = form.locator('input:not([type="checkbox"]):not([type="number"])');
    await textInputs.nth(0).fill('E2E Template');
    await textInputs.nth(1).fill('E2E Lunch');
    await form.locator('select').nth(0).selectOption('meals');
    await form.locator('input[type="number"]').fill('2500');
    await textInputs.nth(2).fill('Template created from E2E');

    await form.locator('button.gradient-primary').click();
    await expect.poll(() => createCalls).toBe(1);

    let templateCard = page
      .locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E Template' })
      .first();
    await expect(templateCard).toBeVisible();

    await templateCard.hover();
    await templateCard.locator('div.flex.items-center.gap-2.mt-3 button').nth(0).click();
    await expect.poll(() => useCalls).toBe(1);
    await expect(page).toHaveURL('/expenses/new');

    await page.goto('/expenses/templates');
    await expect(page).toHaveURL('/expenses/templates');

    templateCard = page
      .locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E Template' })
      .first();
    await templateCard.hover();
    await templateCard.locator('div.flex.items-center.gap-2.mt-3 button').nth(1).click();

    const editForm = page.locator('div.glass-card.rounded-2xl.p-6.space-y-4').first();
    const editTextInputs = editForm.locator('input:not([type="checkbox"]):not([type="number"])');
    await editTextInputs.nth(0).fill('E2E Template Updated');
    await editForm.locator('input[type="number"]').fill('3000');
    await editForm.locator('button.gradient-primary').click();
    await expect.poll(() => updateCalls).toBe(1);
    await expect(page.locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E Template Updated' })).toHaveCount(1);

    const updatedCard = page
      .locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E Template Updated' })
      .first();
    await updatedCard.hover();
    await updatedCard.locator('div.flex.items-center.gap-2.mt-3 button').nth(2).click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page.locator('div.glass-subtle.rounded-xl.p-4', { hasText: 'E2E Template Updated' })).toHaveCount(0);
  });
});
