import { test, expect, type Page } from '@playwright/test';

type AttendanceState = {
  isClockedIn: boolean;
  isClockedOut: boolean;
};

type ApiCalls = {
  today: number;
  list: number;
  summary: number;
  clockIn: number;
  clockOut: number;
};

function buildTodayResponse(state: AttendanceState) {
  if (!state.isClockedIn) return {};
  if (!state.isClockedOut) {
    return { clock_in: '2026-02-12T09:00:00Z' };
  }
  return {
    clock_in: '2026-02-12T09:00:00Z',
    clock_out: '2026-02-12T18:00:00Z',
  };
}

function buildListResponse(state: AttendanceState) {
  return {
    data: [
      {
        id: 'att-1',
        date: '2026-02-12',
        clock_in: state.isClockedIn ? '2026-02-12T09:00:00Z' : null,
        clock_out: state.isClockedOut ? '2026-02-12T18:00:00Z' : null,
        work_minutes: state.isClockedOut ? 540 : 0,
        overtime_minutes: state.isClockedOut ? 60 : 0,
        status: state.isClockedIn ? 'present' : 'absent',
      },
    ],
    total: 1,
    total_pages: 1,
  };
}

function buildSummaryResponse(state: AttendanceState) {
  if (!state.isClockedOut) {
    return {
      total_work_days: state.isClockedIn ? 1 : 0,
      total_work_minutes: 0,
      total_overtime_minutes: 0,
      average_work_minutes: 0,
    };
  }

  return {
    total_work_days: 1,
    total_work_minutes: 540,
    total_overtime_minutes: 60,
    average_work_minutes: 540,
  };
}

async function mockAttendanceFlowApis(page: Page) {
  const state: AttendanceState = {
    isClockedIn: false,
    isClockedOut: false,
  };

  const calls: ApiCalls = {
    today: 0,
    list: 0,
    summary: 0,
    clockIn: 0,
    clockOut: 0,
  };

  await page.route('**/api/v1/**', async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    if (url.endsWith('/auth/login') && method === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          user: {
            id: 'u-attendance',
            email: 'employee@example.com',
            first_name: 'Taro',
            last_name: 'Kintai',
            role: 'employee',
            is_active: true,
          },
          access_token: 'attendance-token',
          refresh_token: 'attendance-refresh-token',
        }),
      });
      return;
    }

    if (url.includes('/attendance/clock-in') && method === 'POST') {
      calls.clockIn += 1;
      state.isClockedIn = true;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ clock_in: '2026-02-12T09:00:00Z' }),
      });
      return;
    }

    if (url.includes('/attendance/clock-out') && method === 'POST') {
      calls.clockOut += 1;
      state.isClockedOut = true;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ clock_out: '2026-02-12T18:00:00Z' }),
      });
      return;
    }

    if (url.includes('/attendance/today') && method === 'GET') {
      calls.today += 1;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildTodayResponse(state)),
      });
      return;
    }

    if (url.includes('/attendance/summary') && method === 'GET') {
      calls.summary += 1;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSummaryResponse(state)),
      });
      return;
    }

    if (url.includes('/attendance?') && method === 'GET') {
      calls.list += 1;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildListResponse(state)),
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

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({}),
    });
  });

  return { calls };
}

async function login(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('employee@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Attendance E2E', () => {
  test('should allow clock-in then clock-out and refetch attendance data', async ({ page }) => {
    const { calls } = await mockAttendanceFlowApis(page);
    await login(page);

    await page.goto('/attendance');
    await expect(page).toHaveURL('/attendance');

    const clockInButton = page.locator('button:has(svg.lucide-log-in)');
    const clockOutButton = page.locator('button:has(svg.lucide-log-out)');

    await expect(clockInButton).toBeVisible();
    await expect(clockOutButton).toBeVisible();
    await expect(clockInButton).toBeEnabled();
    await expect(clockOutButton).toBeDisabled();

    await clockInButton.click();

    await expect.poll(() => calls.clockIn).toBe(1);
    await expect.poll(() => calls.today).toBeGreaterThan(1);
    await expect.poll(() => calls.summary).toBeGreaterThan(1);
    await expect.poll(() => calls.list).toBeGreaterThan(1);
    await expect(clockInButton).toBeDisabled();
    await expect(clockOutButton).toBeEnabled();

    await clockOutButton.click();

    await expect.poll(() => calls.clockOut).toBe(1);
    await expect.poll(() => calls.today).toBeGreaterThan(2);
    await expect.poll(() => calls.summary).toBeGreaterThan(2);
    await expect.poll(() => calls.list).toBeGreaterThan(2);
    await expect(clockOutButton).toBeDisabled();
  });
});
