import { test, expect, type Page } from '@playwright/test';

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Dashboard E2E', () => {
  test('should render dashboard stats and department summary', async ({ page }) => {
    test.setTimeout(90_000);

    let statsCalls = 0;

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
          body: JSON.stringify({
            clock_in: '2026-03-15T09:00:00Z',
          }),
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

      if (path.endsWith('/dashboard/stats') && method === 'GET') {
        statsCalls += 1;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            today_present_count: 12,
            today_absent_count: 1,
            pending_leaves: 2,
            monthly_overtime: 540,
            weekly_trend: [
              { date: '2026-03-09', present_count: 11, absent_count: 2, attendance_rate: 84.6 },
              { date: '2026-03-10', present_count: 12, absent_count: 1, attendance_rate: 92.3 },
            ],
            department_stats: [
              { department_name: 'Engineering', total_employees: 8, present_today: 7, attendance_rate: 0.875 },
              { department_name: 'HR', total_employees: 5, present_today: 5, attendance_rate: 1.0 },
            ],
          }),
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
    await page.goto('/dashboard');
    await expect(page).toHaveURL('/dashboard');
    await expect.poll(() => statsCalls).toBeGreaterThan(0);

    await expect(page.getByText('12').first()).toBeVisible();
    await expect(page.getByText('9h').first()).toBeVisible();

    const departmentTable = page.locator('table:visible').first();
    await expect(departmentTable).toContainText('Engineering');
    await expect(departmentTable).toContainText('HR');
    await expect(departmentTable).toContainText('87.5%');
  });
});
