import { test, expect, type Page } from '@playwright/test';

type ExpenseNotification = {
  id: string;
  type: string;
  title: string;
  message: string;
  is_read: boolean;
  created_at: string;
  expense_id?: string;
};

type Reminder = {
  id: string;
  type: string;
  title: string;
  message: string;
  action_url?: string;
};

type NotificationSettings = {
  on_approved: boolean;
  on_rejected: boolean;
  on_comment: boolean;
  on_reimbursed: boolean;
  month_end_reminder: boolean;
  overdue_reminder: boolean;
  reminder_days_before: number;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Expense Notifications E2E', () => {
  test('should read notifications, dismiss reminder, and update settings', async ({ page }) => {
    test.setTimeout(90_000);

    let notifications: ExpenseNotification[] = [
      {
        id: 'en-1',
        type: 'pending',
        title: 'Pending review',
        message: 'Expense request waiting for approval.',
        is_read: false,
        created_at: '2026-03-10T09:00:00Z',
        expense_id: 'exp-1',
      },
      {
        id: 'en-2',
        type: 'approved',
        title: 'Approved expense',
        message: 'Your expense has been approved.',
        is_read: false,
        created_at: '2026-03-10T10:00:00Z',
      },
      {
        id: 'en-3',
        type: 'comment',
        title: 'Comment added',
        message: 'Approver left a comment.',
        is_read: true,
        created_at: '2026-03-10T11:00:00Z',
      },
    ];

    let reminders: Reminder[] = [
      {
        id: 'rem-1',
        type: 'overdue',
        title: 'Overdue receipt',
        message: 'Please submit receipt by end of week.',
      },
    ];

    let settings: NotificationSettings = {
      on_approved: true,
      on_rejected: true,
      on_comment: true,
      on_reimbursed: true,
      month_end_reminder: true,
      overdue_reminder: true,
      reminder_days_before: 3,
    };

    let markReadCalls = 0;
    let markAllCalls = 0;
    let dismissCalls = 0;
    let settingsUpdateCalls = 0;
    const filtersSeen = new Set<string>();

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

      if (path.endsWith('/expenses/notifications') && method === 'GET') {
        const filter = parsed.searchParams.get('filter') || 'all';
        filtersSeen.add(filter);
        const data =
          filter === 'unread'
            ? notifications.filter((n) => !n.is_read)
            : filter === 'action_required'
              ? notifications.filter((n) => n.type === 'pending' || n.type === 'policy_violation')
              : notifications;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data }),
        });
        return;
      }

      if (/\/expenses\/notifications\/[^/]+\/read$/.test(path) && method === 'PUT') {
        markReadCalls += 1;
        const id = path.split('/')[path.split('/').length - 2];
        notifications = notifications.map((n) => (n.id === id ? { ...n, is_read: true } : n));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (path.endsWith('/expenses/notifications/read-all') && method === 'PUT') {
        markAllCalls += 1;
        notifications = notifications.map((n) => ({ ...n, is_read: true }));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (path.endsWith('/expenses/reminders') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: reminders }),
        });
        return;
      }

      if (/\/expenses\/reminders\/[^/]+\/dismiss$/.test(path) && method === 'PUT') {
        dismissCalls += 1;
        const id = path.split('/')[path.split('/').length - 2];
        reminders = reminders.filter((r) => r.id !== id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      if (path.endsWith('/expenses/notification-settings') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(settings),
        });
        return;
      }

      if (path.endsWith('/expenses/notification-settings') && method === 'PUT') {
        settingsUpdateCalls += 1;
        const body = req.postDataJSON() as NotificationSettings;
        settings = { ...settings, ...body };
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(settings),
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
    await page.goto('/expenses/notifications');
    await expect(page).toHaveURL('/expenses/notifications');

    const pendingRow = page.locator('div.rounded-xl.p-4', { hasText: 'Pending review' }).first();
    await expect(pendingRow).toBeVisible();
    await pendingRow.click();
    await expect.poll(() => markReadCalls).toBe(1);

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("done_all")):visible')
      .first()
      .click();
    await expect.poll(() => markAllCalls).toBe(1);

    await page.locator('div.flex.gap-2.flex-wrap button').nth(1).click();
    await expect.poll(() => filtersSeen.has('unread')).toBeTruthy();

    const reminderCard = page.locator('div.glass-subtle.rounded-xl.p-3', { hasText: 'Overdue receipt' }).first();
    await expect(reminderCard).toBeVisible();
    await reminderCard
      .locator('button:has(span.material-symbols-outlined:has-text("close"))')
      .first()
      .click();
    await expect.poll(() => dismissCalls).toBe(1);
    await expect(page.locator('div.glass-subtle.rounded-xl.p-3', { hasText: 'Overdue receipt' })).toHaveCount(0);

    const firstToggle = page.locator('div.relative.w-11.h-6.rounded-full').first();
    await firstToggle.click();
    await expect.poll(() => settingsUpdateCalls).toBe(1);
    expect(settings.on_approved).toBe(false);
  });
});
