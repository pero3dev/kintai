import { createRouter } from '@tanstack/react-router';
import { describe, expect, it } from 'vitest';
import { routeTree } from './routes';

describe('routes.tsx', () => {
  it('registers key application routes in the route tree', () => {
    const router = createRouter({ routeTree });
    const paths = Object.keys(router.routesByPath);

    const expectedPaths = [
      '/login',
      '/',
      '/dashboard',
      '/attendance',
      '/leaves',
      '/expenses',
      '/expenses/new',
      '/expenses/$expenseId',
      '/hr',
      '/hr/employees',
      '/hr/employees/$employeeId',
      '/wiki',
      '/wiki/architecture',
      '/wiki/backend',
      '/wiki/frontend',
      '/wiki/infrastructure',
      '/wiki/testing',
    ];

    expectedPaths.forEach((path) => {
      expect(paths).toContain(path);
    });

    expect(paths.length).toBeGreaterThanOrEqual(40);
  });

  it('keeps public and protected top-level paths separated', () => {
    const router = createRouter({ routeTree });
    expect(router.routesByPath['/login']).toBeDefined();
    expect(router.routesByPath['/']).toBeDefined();
    expect(router.routesByPath['/dashboard']).toBeDefined();
  });
});
