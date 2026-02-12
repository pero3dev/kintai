import { test, expect, type Page } from '@playwright/test';

type UserRole = 'employee' | 'manager' | 'admin';

type User = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: UserRole;
  is_active: boolean;
};

type Department = {
  id: string;
  name: string;
};

type Employee = {
  id: string;
  employee_id: string;
  first_name: string;
  last_name: string;
  email: string;
  department: string;
  department_name: string;
  position: string;
  employment_type: string;
  hire_date: string;
  phone: string;
  address: string;
  status: string;
};

type ApiCalls = {
  stats: number;
  activities: number;
  createEmployee: number;
  updateEmployee: number;
  goals: number;
  documents: number;
  salaryHistory: number;
};

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('HR Main Flow E2E', () => {
  test('should cover dashboard, employee create/update, and detail tabs', async ({ page }) => {
    test.setTimeout(90_000);

    const departments: Department[] = [
      { id: 'dept-eng', name: 'Engineering' },
      { id: 'dept-hr', name: 'HR' },
    ];

    let employees: Employee[] = [
      {
        id: 'emp-1',
        employee_id: 'EMP-001',
        first_name: 'Taro',
        last_name: 'Yamada',
        email: 'taro.yamada@example.com',
        department: 'dept-eng',
        department_name: 'Engineering',
        position: 'Software Engineer',
        employment_type: 'fullTime',
        hire_date: '2024-04-01',
        phone: '090-1111-1111',
        address: 'Tokyo Minato 1-1',
        status: 'active',
      },
    ];

    const goalsByEmployee: Record<string, Array<Record<string, unknown>>> = {
      'emp-1': [
        {
          id: 'goal-1',
          title: 'Improve onboarding speed',
          progress: 50,
          due_date: '2026-06-30',
        },
      ],
    };

    const documentsByEmployee: Record<string, Array<Record<string, unknown>>> = {
      'emp-1': [
        {
          id: 'doc-1',
          name: 'Employment Contract',
          type: 'contract',
          upload_date: '2025-01-10',
        },
      ],
    };

    const salaryHistoryByEmployee: Record<string, Array<Record<string, unknown>>> = {
      'emp-1': [
        {
          id: 'salary-1',
          effective_date: '2025-04-01',
          base_salary: 350000,
          allowances: 30000,
          deductions: 20000,
          net_salary: 360000,
          reason: 'Annual review raise',
        },
      ],
    };

    const calls: ApiCalls = {
      stats: 0,
      activities: 0,
      createEmployee: 0,
      updateEmployee: 0,
      goals: 0,
      documents: 0,
      salaryHistory: 0,
    };

    const managerUser: User = {
      id: 'u-manager',
      email: 'manager@example.com',
      first_name: 'Jiro',
      last_name: 'Manager',
      role: 'manager',
      is_active: true,
    };

    await page.route('**/api/v1/**', async (route) => {
      const request = route.request();
      const url = request.url();
      const method = request.method();
      const parsed = new URL(url);
      const path = parsed.pathname;

      if (path.endsWith('/auth/login') && method === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            user: managerUser,
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

      if (path.endsWith('/hr/stats') && method === 'GET') {
        calls.stats += 1;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_employees: employees.length,
            active_employees: employees.filter((e) => e.status === 'active').length,
            new_hires_this_month: 1,
            turnover_rate: 2.5,
            open_positions: 2,
            upcoming_reviews: 3,
            training_completion: 80,
            pending_documents: 1,
          }),
        });
        return;
      }

      if (path.endsWith('/hr/activities') && method === 'GET') {
        calls.activities += 1;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [
              {
                icon: 'person_add',
                message: 'New employee registered',
                timestamp: '2026-02-12 09:00',
              },
            ],
          }),
        });
        return;
      }

      if (path.endsWith('/hr/departments') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: departments }),
        });
        return;
      }

      if (path.endsWith('/hr/employees') && method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: employees, total: employees.length }),
        });
        return;
      }

      if (path.endsWith('/hr/employees') && method === 'POST') {
        calls.createEmployee += 1;
        const body = request.postDataJSON() as Record<string, unknown>;
        const newID = `emp-${employees.length + 1}`;
        const departmentID = String(body.department || '');
        const departmentName = departments.find((d) => d.id === departmentID)?.name || '-';

        const created: Employee = {
          id: newID,
          employee_id: String(body.employee_id || ''),
          first_name: String(body.first_name || ''),
          last_name: String(body.last_name || ''),
          email: String(body.email || ''),
          department: departmentID,
          department_name: departmentName,
          position: String(body.position || ''),
          employment_type: String(body.employment_type || 'fullTime'),
          hire_date: String(body.hire_date || ''),
          phone: String(body.phone || ''),
          address: '',
          status: String(body.status || 'active'),
        };

        employees = [...employees, created];
        goalsByEmployee[newID] = [
          {
            id: 'goal-new-1',
            title: 'Automate smoke tests',
            progress: 60,
            due_date: '2026-08-31',
          },
        ];
        documentsByEmployee[newID] = [
          {
            id: 'doc-new-1',
            name: 'Employment Contract',
            type: 'contract',
            upload_date: '2026-03-01',
          },
        ];
        salaryHistoryByEmployee[newID] = [
          {
            id: 'salary-new-1',
            effective_date: '2026-04-01',
            base_salary: 380000,
            allowances: 25000,
            deductions: 18000,
            net_salary: 387000,
            reason: 'Annual review raise',
          },
        ];

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (/\/hr\/employees\/[^/]+$/.test(path) && method === 'GET') {
        const employeeID = path.split('/').pop() || '';
        const target = employees.find((e) => e.id === employeeID);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: target || null }),
        });
        return;
      }

      if (/\/hr\/employees\/[^/]+$/.test(path) && method === 'PUT') {
        calls.updateEmployee += 1;
        const employeeID = path.split('/').pop() || '';
        const body = request.postDataJSON() as Record<string, unknown>;
        employees = employees.map((e) =>
          e.id === employeeID
            ? {
                ...e,
                first_name: String(body.first_name ?? e.first_name),
                last_name: String(body.last_name ?? e.last_name),
                email: String(body.email ?? e.email),
                phone: String(body.phone ?? e.phone),
                position: String(body.position ?? e.position),
                address: String(body.address ?? e.address),
              }
            : e,
        );
        const updated = employees.find((e) => e.id === employeeID) || null;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(updated),
        });
        return;
      }

      if (path.endsWith('/hr/goals') && method === 'GET') {
        calls.goals += 1;
        const employeeID = parsed.searchParams.get('employee_id') || '';
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: goalsByEmployee[employeeID] || [] }),
        });
        return;
      }

      if (path.endsWith('/hr/documents') && method === 'GET') {
        calls.documents += 1;
        const employeeID = parsed.searchParams.get('employee_id') || '';
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: documentsByEmployee[employeeID] || [] }),
        });
        return;
      }

      if (/\/hr\/salary\/[^/]+\/history$/.test(path) && method === 'GET') {
        calls.salaryHistory += 1;
        const segments = path.split('/');
        const employeeID = segments[segments.length - 2] || '';
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: salaryHistoryByEmployee[employeeID] || [] }),
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

    await page.goto('/hr');
    await expect(page).toHaveURL('/hr');
    await expect.poll(() => calls.stats).toBeGreaterThan(0);
    await expect.poll(() => calls.activities).toBeGreaterThan(0);

    await page.goto('/hr/employees');
    await expect(page).toHaveURL('/hr/employees');

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("person_add"))')
      .first()
      .click();

    const addModal = page.locator('div.fixed.inset-0').last();
    await expect(addModal).toBeVisible();

    await addModal.locator('input[placeholder="EMP-001"]').fill('EMP-900');
    await addModal.locator('input[type="email"]').fill('new.employee@example.com');
    const addTextInputs = addModal.locator('input:not([type="email"]):not([type="date"])');
    await addTextInputs.nth(1).fill('E2ELast');
    await addTextInputs.nth(2).fill('E2EFirst');
    await addModal.locator('select').nth(0).selectOption('dept-eng');
    await addTextInputs.nth(3).fill('QA Engineer');
    await addModal.locator('select').nth(1).selectOption('fullTime');
    await addModal.locator('input[type="date"]').fill('2026-03-01');
    await addTextInputs.nth(4).fill('090-0000-0000');

    await addModal.locator('button.gradient-primary').click();
    await expect.poll(() => calls.createEmployee).toBe(1);

    const employeeDetailLink = page
      .locator('a[href^="/hr/employees/"]:visible', { hasText: 'E2ELast E2EFirst' })
      .first();
    await expect(employeeDetailLink).toBeVisible();
    await employeeDetailLink.click();
    await expect(page).toHaveURL(/\/hr\/employees\/emp-\d+$/);

    await page
      .locator('button:has(span.material-symbols-outlined:has-text("edit"))')
      .first()
      .click();

    const editModal = page.locator('div.fixed.inset-0').last();
    await expect(editModal).toBeVisible();

    const editTextInputs = editModal.locator('input:not([type="email"]):not([type="date"])');
    await editTextInputs.nth(2).fill('090-9999-9999');
    await editTextInputs.nth(3).fill('Senior QA Engineer');
    await editTextInputs.nth(4).fill('Tokyo Chiyoda 1-1');

    await editModal.locator('button.gradient-primary').click();
    await expect.poll(() => calls.updateEmployee).toBe(1);
    await expect(page.getByText('Senior QA Engineer').first()).toBeVisible();

    const tabButtons = page
      .locator('div.flex.gap-1.glass-card.rounded-2xl.overflow-x-auto')
      .first()
      .locator('button');

    await tabButtons.nth(1).click();
    await expect.poll(() => calls.goals).toBeGreaterThan(0);
    await expect(page.getByText('Automate smoke tests').first()).toBeVisible();

    await tabButtons.nth(2).click();
    await expect.poll(() => calls.documents).toBeGreaterThan(0);
    await expect(page.getByText('Employment Contract').first()).toBeVisible();

    await tabButtons.nth(3).click();
    await expect.poll(() => calls.salaryHistory).toBeGreaterThan(0);
    await expect(page.getByText('Annual review raise').first()).toBeVisible();
  });
});
