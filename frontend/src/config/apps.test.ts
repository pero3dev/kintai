import { describe, expect, it } from 'vitest';
import { apps, getActiveApp, getAvailableApps } from './apps';

describe('apps config', () => {
  it('returns the most specific enabled app for a pathname', () => {
    expect(getActiveApp('/expenses/reports/2026')?.id).toBe('expenses');
    expect(getActiveApp('/hr/employees')?.id).toBe('hr');
    expect(getActiveApp('/wiki/testing')?.id).toBe('wiki');
  });

  it('ignores disabled apps and falls back to the root app', () => {
    expect(getActiveApp('/tasks/board')?.id).toBe('attendance');
  });

  it('returns undefined when pathname does not match any base path', () => {
    expect(getActiveApp('invalid-path')).toBeUndefined();
  });

  it('includes role-restricted apps for allowed roles', () => {
    const adminAppIds = getAvailableApps('admin').map((app) => app.id);

    expect(adminAppIds).toContain('hr');
    expect(adminAppIds).toHaveLength(apps.length);
  });

  it('excludes role-restricted apps for disallowed or missing roles', () => {
    const employeeAppIds = getAvailableApps('employee').map((app) => app.id);
    const noRoleAppIds = getAvailableApps().map((app) => app.id);

    expect(employeeAppIds).not.toContain('hr');
    expect(noRoleAppIds).not.toContain('hr');
  });
});

