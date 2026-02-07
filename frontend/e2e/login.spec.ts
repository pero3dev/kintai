import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test('should display login form', async ({ page }) => {
    await page.goto('/login');

    // ログインフォームの表示を確認
    await expect(page.locator('form')).toBeVisible();
    await expect(page.getByRole('button')).toBeVisible();
  });

  test('should have input fields', async ({ page }) => {
    await page.goto('/login');

    // 入力フィールドの存在を確認
    const inputs = page.locator('input');
    await expect(inputs).toHaveCount(2);
  });
});
