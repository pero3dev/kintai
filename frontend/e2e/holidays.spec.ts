import { test, expect, type Page } from '@playwright/test';

type Holiday = {
  id: string;
  date: string;
  name: string;
  holiday_type: 'national' | 'company' | 'optional';
  is_recurring?: boolean;
};

function buildCalendarDays(year: number, month: number, holidays: Holiday[]) {
  const daysInMonth = new Date(year, month, 0).getDate();
  const holidayMap = new Map(holidays.map((h) => [h.date, h]));
  const result: Array<Record<string, unknown>> = [];

  for (let day = 1; day <= daysInMonth; day += 1) {
    const date = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
    const weekday = new Date(year, month - 1, day).getDay();
    const holiday = holidayMap.get(date);
    result.push({
      date,
      is_holiday: Boolean(holiday),
      holiday_name: holiday?.name || '',
      is_weekend: weekday === 0 || weekday === 6,
    });
  }

  return result;
}

function buildWorkingDays(startDate: string, endDate: string, holidays: Holiday[]) {
  const start = new Date(startDate);
  const end = new Date(endDate);
  let totalDays = 0;
  let weekends = 0;
  let holidayCount = 0;
  let workingDays = 0;

  const holidaySet = new Set(holidays.map((h) => h.date));

  for (
    let current = new Date(start);
    current <= end;
    current.setDate(current.getDate() + 1)
  ) {
    totalDays += 1;
    const day = current.getDay();
    const date = `${current.getFullYear()}-${String(current.getMonth() + 1).padStart(2, '0')}-${String(current.getDate()).padStart(2, '0')}`;
    const isWeekend = day === 0 || day === 6;
    const isHoliday = holidaySet.has(date);

    if (isWeekend) {
      weekends += 1;
      continue;
    }
    if (isHoliday) {
      holidayCount += 1;
      continue;
    }
    workingDays += 1;
  }

  return {
    working_days: workingDays,
    holidays: holidayCount,
    weekends,
    total_days: totalDays,
  };
}

async function loginAsManager(page: Page) {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill('manager@example.com');
  await page.locator('input[type="password"]').fill('password123');
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL('/');
}

test.describe('Holidays E2E', () => {
  test('should create and delete holiday with calendar updates', async ({ page }) => {
    test.setTimeout(90_000);

    const now = new Date();
    const currentYear = now.getFullYear();
    const currentMonth = now.getMonth() + 1;
    const targetDate = `${currentYear}-${String(currentMonth).padStart(2, '0')}-15`;
    const targetName = 'E2E Founders Day';

    let holidays: Holiday[] = [
      {
        id: 'h-1',
        date: `${currentYear}-${String(currentMonth).padStart(2, '0')}-01`,
        name: 'Monthly Kickoff Day',
        holiday_type: 'company',
      },
    ];

    let listCalls = 0;
    let calendarCalls = 0;
    let workingDaysCalls = 0;
    let createCalls = 0;
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

      if (path.endsWith('/holidays') && method === 'GET') {
        listCalls += 1;
        const year = parsed.searchParams.get('year');
        const list = year
          ? holidays.filter((h) => h.date.startsWith(`${year}-`))
          : holidays;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(list),
        });
        return;
      }

      if (path.endsWith('/holidays/calendar') && method === 'GET') {
        calendarCalls += 1;
        const year = Number(parsed.searchParams.get('year'));
        const month = Number(parsed.searchParams.get('month'));
        const data = buildCalendarDays(
          year || currentYear,
          month || currentMonth,
          holidays,
        );
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(data),
        });
        return;
      }

      if (path.endsWith('/holidays/working-days') && method === 'GET') {
        workingDaysCalls += 1;
        const startDate = parsed.searchParams.get('start_date') || `${currentYear}-${String(currentMonth).padStart(2, '0')}-01`;
        const endDate = parsed.searchParams.get('end_date') || `${currentYear}-${String(currentMonth).padStart(2, '0')}-28`;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(buildWorkingDays(startDate, endDate, holidays)),
        });
        return;
      }

      if (path.endsWith('/holidays') && method === 'POST') {
        createCalls += 1;
        const body = req.postDataJSON() as Record<string, unknown>;
        const created: Holiday = {
          id: `h-${holidays.length + 1}`,
          date: String(body.date || ''),
          name: String(body.name || ''),
          holiday_type: (body.holiday_type as Holiday['holiday_type']) || 'national',
          is_recurring: Boolean(body.is_recurring),
        };
        holidays = [...holidays, created];
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(created),
        });
        return;
      }

      if (/\/holidays\/[^/]+$/.test(path) && method === 'DELETE') {
        deleteCalls += 1;
        const id = path.split('/').pop() || '';
        holidays = holidays.filter((h) => h.id !== id);
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
    await page.goto('/holidays');
    await expect(page).toHaveURL('/holidays');
    await expect.poll(() => listCalls).toBeGreaterThan(0);
    await expect.poll(() => calendarCalls).toBeGreaterThan(0);
    await expect.poll(() => workingDaysCalls).toBeGreaterThan(0);

    await page.locator('button:has(svg.lucide-plus):visible').first().click();
    const form = page.locator('form').first();
    await expect(form).toBeVisible();

    await form.locator('input[name="date"]').fill(targetDate);
    await form.locator('input[name="name"]').fill(targetName);
    await form.locator('select[name="holiday_type"]').selectOption('company');
    await form.locator('button[type="submit"]').click();
    await expect.poll(() => createCalls).toBe(1);

    const holidayCard = page
      .locator('div.glass-subtle.rounded-xl', { hasText: targetName })
      .first();
    await expect(holidayCard).toBeVisible();
    await holidayCard.locator('button').first().click();
    await expect.poll(() => deleteCalls).toBe(1);
    await expect(page.locator('div.glass-subtle.rounded-xl', { hasText: targetName })).toHaveCount(0);
  });
});
