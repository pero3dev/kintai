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

type LeaveRecord = {
  id: string;
  user_id: string;
  user: Pick<User, 'first_name' | 'last_name'>;
  leave_type: 'paid' | 'sick' | 'special' | 'half';
  start_date: string;
  end_date: string;
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

test.describe('Leaves E2E', () => {
  test('should complete leave request and approval flow across roles', async ({ page }) => {
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
    const leaves: LeaveRecord[] = [];
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

      if (url.includes('/leaves/pending') && method === 'GET') {
        const pending = leaves.filter((l) => l.status === 'pending');
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

      if (url.match(/\/leaves\/[^/]+\/approve$/) && method === 'PUT') {
        approveCalls += 1;
        const body = request.postDataJSON() as { status?: 'approved' | 'rejected' };
        const idMatch = url.match(/\/leaves\/([^/]+)\/approve$/);
        const leaveID = idMatch?.[1];
        const target = leaves.find((l) => l.id === leaveID);
        if (target && body.status) {
          target.status = body.status;
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: leaveID, status: body.status || 'pending' }),
        });
        return;
      }

      if (url.match(/\/leaves(\?|$)/) && method === 'POST') {
        createCalls += 1;
        const body = request.postDataJSON() as {
          leave_type: 'paid' | 'sick' | 'special' | 'half';
          start_date: string;
          end_date: string;
          reason?: string;
        };
        leaves.push({
          id: `leave-${leaves.length + 1}`,
          user_id: currentUser.id,
          user: {
            first_name: currentUser.first_name,
            last_name: currentUser.last_name,
          },
          leave_type: body.leave_type,
          start_date: body.start_date,
          end_date: body.end_date,
          reason: body.reason || '',
          status: 'pending',
        });

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: leaves[leaves.length - 1].id, status: 'pending' }),
        });
        return;
      }

      if (url.match(/\/leaves(\?|$)/) && method === 'GET') {
        const mine = leaves.filter((l) => l.user_id === currentUser.id);
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

    const leaveReason = 'E2E leave request';

    await loginAs(page, users.employee.email);
    await page.goto('/leaves');
    await expect(page).toHaveURL('/leaves');

    await page.locator('button:has(svg.lucide-plus)').first().click();
    await page.locator('form input[type="text"]').first().fill(leaveReason);
    await page.locator('input[type="date"]').nth(0).fill('2026-02-20');
    await page.locator('input[type="date"]').nth(1).fill('2026-02-21');
    await page.locator('form button[type="submit"]').click();

    await expect.poll(() => createCalls).toBe(1);
    await expect(page.locator('*:visible', { hasText: leaveReason }).first()).toBeVisible();

    await logoutFromLayout(page);

    await loginAs(page, users.manager.email);
    await page.goto('/leaves');
    await expect(page).toHaveURL('/leaves');

    const pendingCard = page.locator('div.glass-subtle:visible', { hasText: leaveReason }).first();
    await expect(pendingCard).toBeVisible();
    await pendingCard.locator('button').first().click();

    await expect.poll(() => approveCalls).toBe(1);
    await expect(page.locator('div.glass-subtle:visible', { hasText: leaveReason })).toHaveCount(0);

    await logoutFromLayout(page);

    await loginAs(page, users.employee.email);
    await page.goto('/leaves');
    await expect(page).toHaveURL('/leaves');

    const historyItem = page
      .locator('tr:visible, div.glass-subtle:visible', { hasText: leaveReason })
      .first();
    await expect(historyItem).toBeVisible();
    await expect(historyItem.locator('span[class*="bg-green-500/20"]')).toBeVisible();
  });
});
