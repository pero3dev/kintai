import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test.describe.configure({ mode: 'serial' });

  test('should display login form', async ({ page }) => {
    await page.goto('/login');

    await expect(page.locator('form')).toBeVisible();
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should login successfully and redirect to home', async ({ page }) => {
    await page.route('**/api/v1/**', async (route) => {
      const url = route.request().url();
      const method = route.request().method();

      if (url.endsWith('/auth/login') && method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: {
              id: 'u-1',
              email: 'employee@example.com',
              first_name: 'Taro',
              last_name: 'Yamada',
              role: 'employee',
              is_active: true,
            },
            access_token: 'access-token',
            refresh_token: 'refresh-token',
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

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    await page.goto('/login');
    await page.locator('input[type="email"]').fill('employee@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();

    await expect(page).toHaveURL('/');
    await expect(page.locator('a[href="/attendance"]:visible').first()).toBeVisible();
  });

  test('should show error when login fails', async ({ page }) => {
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'invalid credentials' }),
      });
    });

    await page.goto('/login');
    await page.locator('input[type="email"]').fill('employee@example.com');
    await page.locator('input[type="password"]').fill('wrong-password');
    await page.locator('button[type="submit"]').click();

    await expect(page).toHaveURL('/login');
    await expect(page.getByText('invalid credentials')).toBeVisible();
  });

  test('should logout and redirect to login', async ({ page }) => {
    await page.route('**/api/v1/**', async (route) => {
      const url = route.request().url();
      const method = route.request().method();

      if (url.endsWith('/auth/login') && method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: {
              id: 'u-logout',
              email: 'employee@example.com',
              first_name: 'Hanako',
              last_name: 'Tanaka',
              role: 'employee',
              is_active: true,
            },
            access_token: 'access-token',
            refresh_token: 'refresh-token',
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

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    await page.goto('/login');
    await page.locator('input[type="email"]').fill('employee@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');

    await page.locator('button:visible').filter({ hasText: /ログアウト|Logout/ }).first().click();

    await expect(page).toHaveURL('/login');
    await expect(page.locator('form')).toBeVisible();

    const isAuthenticated = await page.evaluate(() => {
      const raw = window.localStorage.getItem('kintai-auth');
      if (!raw) return null;
      try {
        return JSON.parse(raw).state?.isAuthenticated ?? null;
      } catch {
        return null;
      }
    });
    expect(isAuthenticated).toBe(false);
  });

  test('should redirect to login when unauthenticated user opens protected route', async ({ page }) => {
    await page.route('**/api/v1/**', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'unauthorized' }),
      });
    });

    await page.goto('/attendance');

    await expect(page).toHaveURL('/login');
    await expect(page.locator('form')).toBeVisible();
  });

  test('should refresh token and retry protected request on 401', async ({ page }) => {
    let refreshCalls = 0;
    let summaryCalls = 0;
    let refreshPayloadMatched = false;
    let retriedWithNewToken = false;

    await page.route('**/api/v1/**', async (route) => {
      const req = route.request();
      const url = req.url();
      const method = req.method();
      const authHeader = req.headers()['authorization'];

      if (url.endsWith('/auth/login') && method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: {
              id: 'u-2',
              email: 'employee@example.com',
              first_name: 'Jiro',
              last_name: 'Suzuki',
              role: 'employee',
              is_active: true,
            },
            access_token: 'expired-token',
            refresh_token: 'refresh-token',
          }),
        });
        return;
      }

      if (url.endsWith('/auth/refresh') && method === 'POST') {
        refreshCalls += 1;
        const body = req.postDataJSON() as { refresh_token?: string };
        refreshPayloadMatched = body.refresh_token === 'refresh-token';

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            access_token: 'new-token',
            refresh_token: 'refresh-token',
          }),
        });
        return;
      }

      if (url.includes('/attendance/summary')) {
        summaryCalls += 1;
        if (summaryCalls === 1 && authHeader === 'Bearer expired-token') {
          await route.fulfill({
            status: 401,
            contentType: 'application/json',
            body: JSON.stringify({ message: 'token expired' }),
          });
          return;
        }
        if (authHeader === 'Bearer new-token') {
          retriedWithNewToken = true;
        }
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_work_days: 0,
            total_work_minutes: 0,
            total_overtime_minutes: 0,
            average_work_minutes: 0,
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

      if (url.includes('/attendance?')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [],
            total: 0,
            total_pages: 0,
          }),
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

    await page.goto('/login');
    await page.locator('input[type="email"]').fill('employee@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');

    await page.goto('/attendance');
    await expect(page).toHaveURL('/attendance');
    await expect.poll(() => refreshCalls).toBe(1);
    await expect.poll(() => refreshPayloadMatched).toBeTruthy();
    await expect.poll(() => retriedWithNewToken).toBeTruthy();
  });

  test('should redirect to login when refresh fails after 401', async ({ page }) => {
    let refreshCalls = 0;
    let summaryCalls = 0;

    await page.route('**/api/v1/**', async (route) => {
      const req = route.request();
      const url = req.url();
      const method = req.method();
      const authHeader = req.headers()['authorization'];

      if (url.endsWith('/auth/login') && method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: {
              id: 'u-3',
              email: 'employee@example.com',
              first_name: 'Saburo',
              last_name: 'Sato',
              role: 'employee',
              is_active: true,
            },
            access_token: 'expired-token-2',
            refresh_token: 'refresh-token-2',
          }),
        });
        return;
      }

      if (url.endsWith('/auth/refresh') && method === 'POST') {
        refreshCalls += 1;
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'refresh expired' }),
        });
        return;
      }

      if (url.includes('/attendance/summary')) {
        summaryCalls += 1;
        if (summaryCalls === 1 && authHeader === 'Bearer expired-token-2') {
          await route.fulfill({
            status: 401,
            contentType: 'application/json',
            body: JSON.stringify({ message: 'token expired' }),
          });
          return;
        }
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_work_days: 0,
            total_work_minutes: 0,
            total_overtime_minutes: 0,
            average_work_minutes: 0,
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

      if (url.includes('/attendance?')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [],
            total: 0,
            total_pages: 0,
          }),
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

    await page.goto('/login');
    await page.locator('input[type="email"]').fill('employee@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');

    await page.goto('/attendance');
    await expect(page).toHaveURL('/login');
    await expect(page.locator('form')).toBeVisible();
    await expect.poll(() => refreshCalls).toBe(1);
  });
});
