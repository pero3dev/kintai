import { test, expect, type Page } from '@playwright/test';

type Shift = {
  id: string;
  user_id: string;
  date: string;
  shift_type: 'morning' | 'day' | 'evening' | 'night' | 'off';
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Shifts E2E', () => {
  test('should create and delete shift from weekly board', async ({ page }) => {
    test.setTimeout(90_000);

    let shifts: Shift[] = [];
    let createCalls = 0;
    let deleteCalls = 0;
    let listCalls = 0;

    await page.route('**/api/v1/**', async (route) => {
      const req = route.request();
      const url = req.url();
      const method = req.method();
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

      if (path.endsWith('/users') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [
              { id: 'u-emp-1', first_name: 'Taro', last_name: 'Employee', role: 'employee' },
              { id: 'u-emp-2', first_name: 'Hanako', last_name: 'Worker', role: 'employee' },
            ],
            total: 2,
            total_pages: 1,
          }),
        });
        return;
      }

      if (path.endsWith('/shifts') && method === 'GET') {
        listCalls += 1;
        const start = parsed.searchParams.get('start_date') || '';
        const end = parsed.searchParams.get('end_date') || '';
        const data = shifts.filter((s) => s.date >= start && s.date <= end);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(data),
        });
        return;
      }

      if (path.endsWith('/shifts') && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const created: Shift = {
          id: `s-${shifts.length + 1}`,
          user_id: String(body.user_id || ''),
          date: String(body.date || ''),
          shift_type: (body.shift_type as Shift['shift_type']) || 'day',
        };
        shifts = [...shifts, created];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (/\/shifts\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const id = path.split('/').pop() || '';
        shifts = shifts.filter((s) => s.id !== id);
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
    await page.goto('/shifts');
    await expect(page).toHaveURL('/shifts');
    await expect.poll(() => listCalls).toBeGreaterThan(0);

    const employeeRow = page.locator('tr:visible', { hasText: 'Employee Taro' }).first();
    await expect(employeeRow).toBeVisible();
    const targetCell = employeeRow.locator('td').nth(1);

    await targetCell.click();
    const modal = page.locator('div.fixed.inset-0:visible').first();
    await expect(modal).toBeVisible();
    await modal.locator('button:has(svg.lucide-plus)').click();

    await expect.poll(() => createCalls).toBe(1);
    await expect(targetCell.locator('span')).toHaveCount(1);

    await targetCell.click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(targetCell.locator('span')).toHaveCount(0);
  });
});
