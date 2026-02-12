import { test, expect, type Page } from '@playwright/test';

type Project = {
  id: string;
  code: string;
  name: string;
  description?: string;
  status: 'active' | 'completed' | 'archived';
  budget_hours?: number;
};

type TimeEntry = {
  id: string;
  project_id: string;
  date: string;
  minutes: number;
  description?: string;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Projects E2E', () => {
  test('should create project and manage time entries', async ({ page }) => {
    test.setTimeout(90_000);

    let projects: Project[] = [
      {
        id: 'prj-1',
        code: 'PRJ-001',
        name: 'Legacy Migration',
        description: 'Existing active project',
        status: 'active',
        budget_hours: 120,
      },
    ];

    let timeEntries: TimeEntry[] = [];
    let createProjectCalls = 0;
    let createTimeEntryCalls = 0;
    let deleteTimeEntryCalls = 0;
    let summaryCalls = 0;

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

      if (path.endsWith('/projects') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: projects,
            total: projects.length,
            total_pages: 1,
            page: 1,
            page_size: 12,
          }),
        });
        return;
      }

      if (path.endsWith('/projects') && method === 'POST') {
        createProjectCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const project: Project = {
          id: `prj-${projects.length + 1}`,
          code: String(body.code || ''),
          name: String(body.name || ''),
          description: String(body.description || ''),
          status: 'active',
          budget_hours: Number(body.budget_hours || 0) || undefined,
        };
        projects = [...projects, project];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(project),
        });
        return;
      }

      if (path.endsWith('/time-entries') && method === 'GET') {
        const joined = timeEntries.map((entry) => {
          const project = projects.find((p) => p.id === entry.project_id);
          return {
            ...entry,
            project: project ? { id: project.id, name: project.name, code: project.code } : null,
          };
        });
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(joined),
        });
        return;
      }

      if (path.endsWith('/time-entries/summary') && method === 'GET') {
        summaryCalls += 1;
        const summary = projects.map((project) => {
          const entries = timeEntries.filter((entry) => entry.project_id === project.id);
          const totalMinutes = entries.reduce((acc, entry) => acc + entry.minutes, 0);
          const memberCount = entries.length > 0 ? 1 : 0;
          return {
            project_code: project.code,
            project_name: project.name,
            total_hours: totalMinutes / 60,
            budget_hours: project.budget_hours || null,
            member_count: memberCount,
          };
        });
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(summary),
        });
        return;
      }

      if (path.endsWith('/time-entries') && method === 'POST') {
        createTimeEntryCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const entry: TimeEntry = {
          id: `te-${timeEntries.length + 1}`,
          project_id: String(body.project_id || ''),
          date: String(body.date || ''),
          minutes: Number(body.minutes || 0),
          description: String(body.description || ''),
        };
        timeEntries = [...timeEntries, entry];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(entry),
        });
        return;
      }

      if (/\/time-entries\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteTimeEntryCalls += 1;
        const entryID = path.split('/').pop() || '';
        timeEntries = timeEntries.filter((entry) => entry.id !== entryID);
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
    await page.goto('/projects');
    await expect(page).toHaveURL('/projects');

    await page.locator('button:has(svg.lucide-plus):visible').first().click();
    const projectForm = page
      .locator('form')
      .filter({ has: page.locator('input[placeholder="PRJ-001"]') })
      .first();
    await expect(projectForm).toBeVisible();

    await projectForm.locator('input[type="text"]').nth(0).fill('E2E Project');
    await projectForm.locator('input[placeholder="PRJ-001"]').fill('PRJ-E2E');
    await projectForm.locator('input[type="text"]').nth(2).fill('E2E description');
    await projectForm.locator('input[type="number"]').fill('80');
    await projectForm.locator('button[type="submit"]').click();

    await expect.poll(() => createProjectCalls).toBe(1);
    await expect(page.getByText('PRJ-E2E').first()).toBeVisible();
    await expect(page.getByText('E2E Project').first()).toBeVisible();

    await page.locator('button:has(svg.lucide-clock):visible').first().click();
    const timeForm = page
      .locator('form')
      .filter({ has: page.locator('input[type="date"]') })
      .first();
    await expect(timeForm).toBeVisible();

    await timeForm.locator('select').selectOption({ label: 'E2E Project' });
    await timeForm.locator('input[type="date"]').fill('2026-03-15');
    await timeForm.locator('input[type="number"]').fill('90');
    await timeForm.locator('input[type="text"]').fill('E2E implementation work');
    await timeForm.locator('button[type="submit"]').click();

    await expect.poll(() => createTimeEntryCalls).toBe(1);

    const tabs = page.locator('div.flex.gap-1.border-b.border-border button');
    await tabs.nth(1).click();

    const entryRow = page.locator('tbody tr:visible', { hasText: 'E2E implementation work' }).first();
    await expect(entryRow).toBeVisible();
    await expect(entryRow).toContainText('90');

    await entryRow.locator('button').first().click();
    await expect.poll(() => deleteTimeEntryCalls).toBe(1);
    await expect(page.locator('tbody tr:visible', { hasText: 'E2E implementation work' })).toHaveCount(0);

    await tabs.nth(2).click();
    await expect.poll(() => summaryCalls).toBeGreaterThan(0);
    await expect(page.getByText('PRJ-E2E').first()).toBeVisible();
  });
});
