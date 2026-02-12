import { test, expect, type Page } from '@playwright/test';

type User = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: 'employee';
  is_active: boolean;
};

type ExpenseItem = {
  expense_date: string;
  category: string;
  description: string;
  amount: number;
  receipt_url?: string;
};

type ExpenseRecord = {
  id: string;
  user_id: string;
  user_name: string;
  title: string;
  status: 'draft' | 'pending' | 'approved' | 'rejected';
  notes: string;
  created_at: string;
  amount: number;
  expense_date: string;
  category: string;
  items: ExpenseItem[];
};

type Comment = {
  id: string;
  user_name: string;
  content: string;
  created_at: string;
};

type HistoryEntry = {
  action: string;
  details: string;
  user_name: string;
  created_at: string;
};

async function loginAsEmployee(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('employee@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Expenses History/Detail E2E', () => {
  test('should add comment, submit draft, and delete draft from detail page', async ({ page }) => {
    test.setTimeout(90_000);

    const user: User = {
      id: 'u-employee',
      email: 'employee@example.com',
      first_name: 'Taro',
      last_name: 'Employee',
      role: 'employee',
      is_active: true,
    };

    const submitTitle = 'E2E Draft Submit';
    const deleteTitle = 'E2E Draft Delete';
    const archivedTitle = 'E2E Approved Archived';
    const submitID = 'exp-1';
    const deleteID = 'exp-2';
    const archivedID = 'exp-3';

    let expenses: ExpenseRecord[] = [
      {
        id: submitID,
        user_id: user.id,
        user_name: `${user.last_name} ${user.first_name}`,
        title: submitTitle,
        status: 'draft',
        notes: 'Will be submitted',
        created_at: '2026-02-26T09:00:00Z',
        amount: 3800,
        expense_date: '2026-02-26',
        category: 'meals',
        items: [
          {
            expense_date: '2026-02-26',
            category: 'meals',
            description: 'Team lunch',
            amount: 3800,
          },
        ],
      },
      {
        id: deleteID,
        user_id: user.id,
        user_name: `${user.last_name} ${user.first_name}`,
        title: deleteTitle,
        status: 'draft',
        notes: 'Will be deleted',
        created_at: '2026-02-27T09:00:00Z',
        amount: 1200,
        expense_date: '2026-02-27',
        category: 'transportation',
        items: [
          {
            expense_date: '2026-02-27',
            category: 'transportation',
            description: 'Taxi',
            amount: 1200,
          },
        ],
      },
      {
        id: archivedID,
        user_id: user.id,
        user_name: `${user.last_name} ${user.first_name}`,
        title: archivedTitle,
        status: 'approved',
        notes: 'Already approved',
        created_at: '2026-02-25T09:00:00Z',
        amount: 2400,
        expense_date: '2026-02-25',
        category: 'supplies',
        items: [
          {
            expense_date: '2026-02-25',
            category: 'supplies',
            description: 'Stationery',
            amount: 2400,
          },
        ],
      },
    ];

    const commentsByExpense: Record<string, Comment[]> = {
      [submitID]: [],
      [deleteID]: [],
      [archivedID]: [],
    };

    const historyByExpense: Record<string, HistoryEntry[]> = {
      [submitID]: [
        {
          action: 'Created',
          details: '',
          user_name: `${user.last_name} ${user.first_name}`,
          created_at: '2026-02-26T09:00:00Z',
        },
      ],
      [deleteID]: [
        {
          action: 'Created',
          details: '',
          user_name: `${user.last_name} ${user.first_name}`,
          created_at: '2026-02-27T09:00:00Z',
        },
      ],
      [archivedID]: [
        {
          action: 'Approved',
          details: '',
          user_name: 'Manager User',
          created_at: '2026-02-25T10:00:00Z',
        },
      ],
    };

    let addCommentCalls = 0;
    let submitCalls = 0;
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
            user,
            access_token: 'access-employee',
            refresh_token: 'refresh-employee',
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

      if (path.endsWith('/expenses') && method === 'GET') {
        let filtered = [...expenses];
        const status = parsed.searchParams.get('status');
        const category = parsed.searchParams.get('category');
        const pageNumber = Number(parsed.searchParams.get('page') || '1');
        const pageSize = Number(parsed.searchParams.get('page_size') || '10');

        if (status) {
          filtered = filtered.filter((e) => e.status === status);
        }
        if (category) {
          filtered = filtered.filter((e) => e.category === category);
        }

        const start = (pageNumber - 1) * pageSize;
        const paged = filtered.slice(start, start + pageSize);

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: paged, total: filtered.length }),
        });
        return;
      }

      if (/\/expenses\/[^/]+\/comments$/.test(path) && method === 'GET') {
        const expenseID = path.split('/').slice(-2)[0] || '';
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: commentsByExpense[expenseID] || [] }),
        });
        return;
      }

      if (/\/expenses\/[^/]+\/comments$/.test(path) && method === 'POST') {
        addCommentCalls += 1;
        const expenseID = path.split('/').slice(-2)[0] || '';
        const body = req.postDataJSON() as { content?: string };
        const nextComment: Comment = {
          id: `c-${(commentsByExpense[expenseID] || []).length + 1}`,
          user_name: `${user.last_name} ${user.first_name}`,
          content: body.content || '',
          created_at: '2026-02-27T10:00:00Z',
        };
        commentsByExpense[expenseID] = [...(commentsByExpense[expenseID] || []), nextComment];
        historyByExpense[expenseID] = [
          ...(historyByExpense[expenseID] || []),
          {
            action: 'Commented',
            details: nextComment.content,
            user_name: nextComment.user_name,
            created_at: nextComment.created_at,
          },
        ];

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(nextComment),
        });
        return;
      }

      if (/\/expenses\/[^/]+\/history$/.test(path) && method === 'GET') {
        const expenseID = path.split('/').slice(-2)[0] || '';
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ data: historyByExpense[expenseID] || [] }),
        });
        return;
      }

      if (/\/expenses\/[^/]+$/.test(path) && method === 'GET') {
        const expenseID = path.split('/').pop() || '';
        const target = expenses.find((e) => e.id === expenseID) || null;
        await route.fulfill({
          status: target ? 200 : 404,
          contentType: 'application/json',
          body: JSON.stringify(target),
        });
        return;
      }

      if (/\/expenses\/[^/]+$/.test(path) && method === 'PUT') {
        submitCalls += 1;
        const expenseID = path.split('/').pop() || '';
        const body = req.postDataJSON() as { status?: ExpenseRecord['status'] };
        expenses = expenses.map((expense) =>
          expense.id === expenseID
            ? {
                ...expense,
                status: body.status || expense.status,
              }
            : expense,
        );
        historyByExpense[expenseID] = [
          ...(historyByExpense[expenseID] || []),
          {
            action: 'Submitted',
            details: 'Status changed to pending',
            user_name: `${user.last_name} ${user.first_name}`,
            created_at: '2026-02-27T11:00:00Z',
          },
        ];
        const updated = expenses.find((expense) => expense.id === expenseID) || null;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(updated),
        });
        return;
      }

      if (/\/expenses\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const expenseID = path.split('/').pop() || '';
        expenses = expenses.filter((expense) => expense.id !== expenseID);
        delete commentsByExpense[expenseID];
        delete historyByExpense[expenseID];
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

    await loginAsEmployee(page);
    await page.goto('/expenses/history');
    await expect(page).toHaveURL('/expenses/history');

    await expect(page.locator('*:visible', { hasText: submitTitle }).first()).toBeVisible();
    await page.locator('select').first().selectOption('draft');
    await expect(page.locator('*:visible', { hasText: archivedTitle })).toHaveCount(0);

    await page.getByRole('link', { name: submitTitle }).first().click();
    await expect(page).toHaveURL(`/expenses/${submitID}`);

    await page.locator('button:has(span.material-symbols-outlined:has-text("chat"))').first().click();
    const commentText = 'E2E comment from detail';
    await page.locator('textarea').first().fill(commentText);
    await page
      .locator('div.glass-subtle.rounded-xl.p-4 button:has(span.material-symbols-outlined:has-text("send"))')
      .first()
      .click();
    await expect.poll(() => addCommentCalls).toBe(1);
    await expect(page.locator('*:visible', { hasText: commentText }).first()).toBeVisible();

    await page.locator('button:has(span.material-symbols-outlined:has-text("history"))').first().click();
    await expect(page.locator('*:visible', { hasText: 'Created' }).first()).toBeVisible();

    await page
      .locator('button.px-4.py-2.gradient-primary:has(span.material-symbols-outlined:has-text("send"))')
      .first()
      .click();
    await expect.poll(() => submitCalls).toBe(1);
    await expect(page.locator('span[class*="bg-yellow-500/20"]').first()).toBeVisible();

    await page.locator('button:has(span.material-symbols-outlined:has-text("arrow_back"))').first().click();
    await expect(page).toHaveURL('/expenses/history');

    await page.locator('select').first().selectOption('pending');
    await expect(page.locator('*:visible', { hasText: submitTitle }).first()).toBeVisible();

    await page.locator('select').first().selectOption('draft');
    await page.getByRole('link', { name: deleteTitle }).first().click();
    await expect(page).toHaveURL(`/expenses/${deleteID}`);

    await page.locator('button:has(span.material-symbols-outlined:has-text("delete"))').first().click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page).toHaveURL('/expenses/history');
    await expect(page.locator('*:visible', { hasText: deleteTitle })).toHaveCount(0);
  });
});
