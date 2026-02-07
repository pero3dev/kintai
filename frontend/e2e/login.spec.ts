import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test('should display login form', async ({ page }) => {
    await page.goto('/login');
    
    await expect(page.getByRole('heading', { name: /ログイン/i })).toBeVisible();
    await expect(page.getByLabel(/メールアドレス/i)).toBeVisible();
    await expect(page.getByLabel(/パスワード/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /ログイン/i })).toBeVisible();
  });

  test('should show error on invalid credentials', async ({ page }) => {
    await page.goto('/login');
    
    await page.getByLabel(/メールアドレス/i).fill('invalid@example.com');
    await page.getByLabel(/パスワード/i).fill('wrongpassword');
    await page.getByRole('button', { name: /ログイン/i }).click();
    
    // Expect error message or toast
    await expect(page.locator('text=/エラー|失敗|認証/')).toBeVisible({ timeout: 5000 });
  });
});
