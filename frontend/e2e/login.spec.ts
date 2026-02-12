import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
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
});
