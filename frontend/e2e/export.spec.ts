import { test, expect, type Page } from '@playwright/test';

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Shared Export E2E', () => {
  test('should export attendance/leaves/overtime/projects as CSV', async ({ page }) => {
    test.setTimeout(60_000);

    const exportCalls: Array<{ type: string; url: string }> = [];

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

      if (path.endsWith('/export/attendance') && method === 'GET') {
        exportCalls.push({ type: 'attendance', url });
        await route.fulfill({
          status: 200,
          contentType: 'text/csv',
          body: 'date,minutes\n2026-03-01,480\n',
        });
        return;
      }

      if (path.endsWith('/export/leaves') && method === 'GET') {
        exportCalls.push({ type: 'leaves', url });
        await route.fulfill({
          status: 200,
          contentType: 'text/csv',
          body: 'date,type\n2026-03-01,paid\n',
        });
        return;
      }

      if (path.endsWith('/export/overtime') && method === 'GET') {
        exportCalls.push({ type: 'overtime', url });
        await route.fulfill({
          status: 200,
          contentType: 'text/csv',
          body: 'date,minutes\n2026-03-01,60\n',
        });
        return;
      }

      if (path.endsWith('/export/projects') && method === 'GET') {
        exportCalls.push({ type: 'projects', url });
        await route.fulfill({
          status: 200,
          contentType: 'text/csv',
          body: 'project,hours\nKINTAI,12\n',
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
    await page.goto('/export');
    await expect(page).toHaveURL('/export');

    await page.locator('input[type="date"]').nth(0).fill('2026-03-01');
    await page.locator('input[type="date"]').nth(1).fill('2026-03-31');

    const exportButtons = page
      .locator('div.grid.grid-cols-1.md\\:grid-cols-2.gap-4')
      .first()
      .locator('button');
    await expect(exportButtons).toHaveCount(4);

    for (let i = 0; i < 4; i += 1) {
      const before = exportCalls.length;
      await exportButtons.nth(i).click();
      await expect.poll(() => exportCalls.length).toBe(before + 1);
    }

    const byType = Object.fromEntries(exportCalls.map((c) => [c.type, c.url]));
    expect(byType.attendance).toContain('/export/attendance?');
    expect(byType.leaves).toContain('/export/leaves?');
    expect(byType.overtime).toContain('/export/overtime?');
    expect(byType.projects).toContain('/export/projects?');

    for (const url of Object.values(byType)) {
      expect(url).toContain('start_date=2026-03-01');
      expect(url).toContain('end_date=2026-03-31');
    }
  });
});
