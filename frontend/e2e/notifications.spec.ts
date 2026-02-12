import { test, expect, type Page } from '@playwright/test';

type Notification = {
  id: string;
  type: string;
  title: string;
  message: string;
  created_at: string;
  is_read: boolean;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Notifications E2E', () => {
  test('should mark one, mark all, and delete notification', async ({ page }) => {
    test.setTimeout(90_000);

    let notifications: Notification[] = [
      {
        id: 'n-1',
        type: 'leave_approved',
        title: 'Leave pending',
        message: 'A leave request needs your review.',
        created_at: '2026-03-01T09:00:00Z',
        is_read: false,
      },
      {
        id: 'n-2',
        type: 'overtime_approved',
        title: 'Overtime approved',
        message: 'Your overtime request was approved.',
        created_at: '2026-03-01T10:00:00Z',
        is_read: false,
      },
      {
        id: 'n-3',
        type: 'general',
        title: 'General info',
        message: 'System maintenance scheduled.',
        created_at: '2026-03-01T11:00:00Z',
        is_read: true,
      },
    ];

    let markOneCalls = 0;
    let markAllCalls = 0;
    let deleteCalls = 0;

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

      if (path.endsWith('/leaves/pending') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (path.endsWith('/notifications') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: notifications,
            total: notifications.length,
            total_pages: 1,
            page: 1,
            page_size: 20,
          }),
        });
        return;
      }

      if (path.endsWith('/notifications/unread-count') && method === 'GET') {
        const unread = notifications.filter((n) => !n.is_read).length;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ unread, count: unread }),
        });
        return;
      }

      if (/\/notifications\/[^/]+\/read$/.test(path) && method === 'PUT') {
        markOneCalls += 1;
        const id = path.split('/')[path.split('/').length - 2];
        notifications = notifications.map((n) =>
          n.id === id ? { ...n, is_read: true } : n,
        );
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (path.endsWith('/notifications/read-all') && method === 'PUT') {
        markAllCalls += 1;
        notifications = notifications.map((n) => ({ ...n, is_read: true }));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (/\/notifications\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const id = path.split('/').pop() || '';
        notifications = notifications.filter((n) => n.id !== id);
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
    await page.goto('/notifications');
    await expect(page).toHaveURL('/notifications');

    const leaveRow = page.locator('div.p-4.flex.items-start.gap-4', { hasText: 'Leave pending' }).first();
    await expect(leaveRow).toBeVisible();
    await leaveRow.locator('button').first().click();
    await expect.poll(() => markOneCalls).toBe(1);

    await page.locator('button:has(svg.lucide-check-check):visible').first().click();
    await expect.poll(() => markAllCalls).toBe(1);
    await expect(page.locator('button:has(svg.lucide-check)')).toHaveCount(0);

    const generalRow = page.locator('div.p-4.flex.items-start.gap-4', { hasText: 'General info' }).first();
    await expect(generalRow).toBeVisible();
    await generalRow.locator('button').first().click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page.locator('div.p-4.flex.items-start.gap-4', { hasText: 'General info' })).toHaveCount(0);
  });
});
