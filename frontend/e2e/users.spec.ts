import { test, expect, type Page } from '@playwright/test';

type Role = 'employee' | 'manager' | 'admin';

type User = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: Role;
  is_active: boolean;
  department_id?: string;
};

type Department = {
  id: string;
  name: string;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Users Management E2E', () => {
  test('should create, update, and delete user', async ({ page }) => {
    test.setTimeout(90_000);

    const departments: Department[] = [
      { id: 'dept-eng', name: 'Engineering' },
      { id: 'dept-hr', name: 'HR' },
    ];

    let users: User[] = [
      {
        id: 'u-admin-1',
        email: 'admin@example.com',
        first_name: 'Admin',
        last_name: 'Root',
        role: 'admin',
        is_active: true,
        department_id: 'dept-hr',
      },
      {
        id: 'u-emp-1',
        email: 'employee.old@example.com',
        first_name: 'Old',
        last_name: 'Employee',
        role: 'employee',
        is_active: true,
        department_id: 'dept-eng',
      },
    ];

    let createCalls = 0;
    let updateCalls = 0;
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

      if (path.endsWith('/notifications') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: [] }),
        });
        return;
      }

      if (path.endsWith('/notifications/unread-count') && method === 'GET') {
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

      if (path.endsWith('/departments') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(departments),
        });
        return;
      }

      if (path.endsWith('/users') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: users,
            total: users.length,
            total_pages: 1,
            page: 1,
            page_size: 20,
          }),
        });
        return;
      }

      if (path.endsWith('/users') && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const newUser: User = {
          id: `u-new-${users.length + 1}`,
          email: String(body.email || ''),
          first_name: String(body.first_name || ''),
          last_name: String(body.last_name || ''),
          role: (body.role as Role) || 'employee',
          is_active: true,
          department_id: String(body.department_id || '') || undefined,
        };
        users = [...users, newUser];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(newUser),
        });
        return;
      }

      if (/\/users\/[^/]+$/.test(path) && method === 'PUT') {
        updateCalls += 1;
        const userID = path.split('/').pop() || '';
        const body = req.postDataJSON() as Record<string, unknown>;
        users = users.map((u) =>
          u.id === userID
            ? {
                ...u,
                first_name: String(body.first_name ?? u.first_name),
                last_name: String(body.last_name ?? u.last_name),
                role: (body.role as Role) ?? u.role,
                is_active: Boolean(body.is_active ?? u.is_active),
                department_id: String(body.department_id ?? u.department_id ?? '') || undefined,
              }
            : u,
        );
        const updated = users.find((u) => u.id === userID) || null;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(updated),
        });
        return;
      }

      if (/\/users\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const userID = path.split('/').pop() || '';
        users = users.filter((u) => u.id !== userID);
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

    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    await loginAsManager(page);
    await page.goto('/users');
    await expect(page).toHaveURL('/users');

    await page
      .locator('button:has(svg.lucide-user-plus):visible')
      .first()
      .click();

    const createModal = page.locator('div.fixed.inset-0:visible').first();
    await expect(createModal).toBeVisible();

    await createModal.locator('input[type="email"]').fill('new.user@example.com');
    await createModal.locator('input[type="password"]').fill('password1234');
    await createModal.locator('input[type="text"]').nth(0).fill('NewLast');
    await createModal.locator('input[type="text"]').nth(1).fill('NewFirst');
    await createModal.locator('select').nth(0).selectOption('employee');
    await createModal.locator('select').nth(1).selectOption('dept-eng');
    await createModal.locator('button:has(svg.lucide-plus)').click();

    await expect.poll(() => createCalls).toBe(1);

    const createdRow = page.locator('tr:visible', { hasText: 'new.user@example.com' }).first();
    await expect(createdRow).toBeVisible();
    await expect(createdRow).toContainText('NewLast NewFirst');

    await createdRow.locator('button:visible').filter({ hasText: /./ }).first().click();

    const editModal = page.locator('div.fixed.inset-0:visible').first();
    await expect(editModal).toBeVisible();
    await editModal.locator('input[type="text"]').nth(0).fill('UpdatedLast');
    await editModal.locator('input[type="text"]').nth(1).fill('UpdatedFirst');
    await editModal.locator('select').nth(0).selectOption('manager');
    await editModal.locator('button:has(svg.lucide-save)').click();

    await expect.poll(() => updateCalls).toBe(1);
    const updatedRow = page.locator('tr:visible', { hasText: 'new.user@example.com' }).first();
    await expect(updatedRow).toContainText('UpdatedLast UpdatedFirst');
    await updatedRow.locator('button[title]:visible').first().click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page.locator('tr:visible', { hasText: 'new.user@example.com' })).toHaveCount(0);
  });
});
