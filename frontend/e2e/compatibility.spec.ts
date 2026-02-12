import { test, expect, type Page } from '@playwright/test';

async function mockAuthAndHomeApis(page: Page) {
  await page.route('**/api/v1/**', async (route) => {
    const req = route.request();
    const url = req.url();
    const method = req.method();

    if (url.endsWith('/auth/login') && method === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          user: {
            id: 'u-compat',
            email: 'compat@example.com',
            first_name: 'Compat',
            last_name: 'User',
            role: 'employee',
            is_active: true,
          },
          access_token: 'access-token-compat',
          refresh_token: 'refresh-token-compat',
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

    if (url.includes('/leaves/pending')) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      });
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({}),
    });
  });
}

async function login(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('compat@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Cross Browser and Device Compatibility', () => {
  test('login screen is rendered', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('form')).toBeVisible();
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('authenticated layout adapts to viewport size', async ({ page }) => {
    await mockAuthAndHomeApis(page);
    await login(page);

    const viewport = page.viewportSize();
    const width = viewport?.width ?? 1280;

    if (width < 768) {
      await expect(page.locator('nav.mobile-bottom-nav')).toBeVisible();
      await expect(page.locator('aside.hidden.md\\:flex')).toBeHidden();
      return;
    }

    await expect(page.locator('nav.mobile-bottom-nav')).toBeHidden();
    await expect(page.locator('aside.hidden.md\\:flex')).toBeVisible();
    await expect(page.locator('a[href="/attendance"]:visible').first()).toBeVisible();
  });
});
