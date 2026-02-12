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

type OvertimeRecord = {
  id: string;
  user_id: string;
  user: Pick<User, 'first_name' | 'last_name'>;
  date: string;
  planned_minutes: number;
  reason: string;
  status: 'pending' | 'approved' | 'rejected';
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

test.describe('Overtime E2E', () => {
  test('should complete overtime request and approval flow across roles', async ({ page }) => {
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

    let currentUser: User = users.employee;
    const overtimes: OvertimeRecord[] = [];
    let createCalls = 0;
    let approveCalls = 0;

    await page.route('**/api/v1/**', async (route) => {
      const request = route.request();
      const url = request.url();
      const method = request.method();

      if (url.endsWith('/auth/login') && method === 'POST') {
        const body = request.postDataJSON() as { email?: string };
        currentUser = body.email === users.manager.email ? users.manager : users.employee;

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: currentUser,
            access_token: `access-token-${currentUser.role}`,
            refresh_token: `refresh-token-${currentUser.role}`,
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

      if (url.includes('/overtime/alerts') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([]),
        });
        return;
      }

      if (url.includes('/overtime/pending') && method === 'GET') {
        const pending = overtimes.filter((o) => o.status === 'pending');
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: pending,
            total: pending.length,
            total_pages: 1,
          }),
        });
        return;
      }

      if (url.match(/\/overtime\/[^/]+\/approve$/) && method === 'PUT') {
        approveCalls += 1;
        const body = request.postDataJSON() as { status?: 'approved' | 'rejected' };
        const idMatch = url.match(/\/overtime\/([^/]+)\/approve$/);
        const overtimeID = idMatch?.[1];
        const target = overtimes.find((o) => o.id === overtimeID);
        if (target && body.status) {
          target.status = body.status;
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: overtimeID, status: body.status || 'pending' }),
        });
        return;
      }

      if (url.match(/\/overtime(\?|$)/) && method === 'POST') {
        createCalls += 1;
        const body = request.postDataJSON() as {
          date: string;
          planned_minutes: number;
          reason?: string;
        };
        overtimes.push({
          id: `ot-${overtimes.length + 1}`,
          user_id: currentUser.id,
          user: {
            first_name: currentUser.first_name,
            last_name: currentUser.last_name,
          },
          date: body.date,
          planned_minutes: body.planned_minutes,
          reason: body.reason || '',
          status: 'pending',
        });

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: overtimes[overtimes.length - 1].id, status: 'pending' }),
        });
        return;
      }

      if (url.match(/\/overtime(\?|$)/) && method === 'GET') {
        const mine = overtimes.filter((o) => o.user_id === currentUser.id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: mine,
            total: mine.length,
            total_pages: 1,
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

    const overtimeReason = 'E2E overtime request';

    await loginAs(page, users.employee.email);
    await page.goto('/overtime');
    await expect(page).toHaveURL('/overtime');

    await page.locator('button:has(svg.lucide-plus)').first().click();
    await page.locator('form input[type="date"]').fill('2026-02-22');
    await page.locator('form input[type="number"]').fill('120');
    await page.locator('form input[type="text"]').fill(overtimeReason);
    await page.locator('form button[type="submit"]').click();

    await expect.poll(() => createCalls).toBe(1);
    await expect(page.locator('*:visible', { hasText: overtimeReason }).first()).toBeVisible();

    await logoutFromLayout(page);

    await loginAs(page, users.manager.email);
    await page.goto('/overtime');
    await expect(page).toHaveURL('/overtime');

    const pendingCard = page.locator('div.glass-subtle:visible', { hasText: overtimeReason }).first();
    await expect(pendingCard).toBeVisible();
    await pendingCard.locator('button').first().click();

    await expect.poll(() => approveCalls).toBe(1);
    await expect(page.locator('div.glass-subtle:visible', { hasText: overtimeReason })).toHaveCount(0);

    await logoutFromLayout(page);

    await loginAs(page, users.employee.email);
    await page.goto('/overtime');
    await expect(page).toHaveURL('/overtime');

    const historyItem = page
      .locator('tr:visible, div.glass-subtle:visible', { hasText: overtimeReason })
      .first();
    await expect(historyItem).toBeVisible();
    await expect(historyItem.locator('span[class*="bg-green-500/20"]')).toBeVisible();
  });
});
