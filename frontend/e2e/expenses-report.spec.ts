import { test, expect, type Page } from '@playwright/test';

async function loginAsEmployee(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('employee@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Expenses Report E2E', () => {
  test('should export expense report as CSV and PDF', async ({ page }) => {
    test.setTimeout(60_000);

    let csvUrl = '';
    let pdfUrl = '';

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
              id: 'u-employee',
              email: 'employee@example.com',
              first_name: 'Taro',
              last_name: 'Employee',
              role: 'employee',
              is_active: true,
            },
            access_token: 'access-employee',
            refresh_token: 'refresh-employee',
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

      if (url.includes('/expenses/report?') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_amount: 18000,
            total_count: 3,
            category_breakdown: [
              { category: 'meals', amount: 12000 },
              { category: 'transportation', amount: 6000 },
            ],
            department_breakdown: [
              { department: 'Engineering', amount: 18000, count: 3 },
            ],
            status_summary: {
              draft: 0,
              pending: 1,
              approved: 2,
              rejected: 0,
              reimbursed: 0,
              approved_amount: 12000,
              pending_amount: 6000,
            },
          }),
        });
        return;
      }

      if (url.includes('/expenses/report/monthly?') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [
              { month: '01', amount: 10000 },
              { month: '02', amount: 8000 },
            ],
          }),
        });
        return;
      }

      if (url.includes('/expenses/export/csv?') && method === 'GET') {
        csvUrl = url;
        await route.fulfill({
          status: 200,
          contentType: 'text/csv',
          body: 'title,amount\nsample,1000\n',
        });
        return;
      }

      if (url.includes('/expenses/export/pdf?') && method === 'GET') {
        pdfUrl = url;
        await route.fulfill({
          status: 200,
          contentType: 'application/pdf',
          body: '%PDF-1.4 mock',
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    await loginAsEmployee(page);
    await page.goto('/expenses/report');
    await expect(page).toHaveURL('/expenses/report');

    const csvDownloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: 'CSV' }).click();
    const csvDownload = await csvDownloadPromise;
    expect(csvDownload.suggestedFilename().endsWith('.csv')).toBeTruthy();
    expect(csvUrl).toContain('/expenses/export/csv?');
    expect(csvUrl).toContain('start_date=');
    expect(csvUrl).toContain('end_date=');

    const pdfDownloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: 'PDF' }).click();
    const pdfDownload = await pdfDownloadPromise;
    expect(pdfDownload.suggestedFilename().endsWith('.pdf')).toBeTruthy();
    expect(pdfUrl).toContain('/expenses/export/pdf?');
    expect(pdfUrl).toContain('start_date=');
    expect(pdfUrl).toContain('end_date=');
  });
});
