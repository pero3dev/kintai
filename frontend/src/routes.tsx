import {
  createRootRoute,
  createRoute,
  Outlet,
} from '@tanstack/react-router';
import { Layout } from './components/layout/Layout';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { AttendancePage } from './pages/AttendancePage';
import { LeavesPage } from './pages/LeavesPage';
import { ShiftsPage } from './pages/ShiftsPage';
import { UsersPage } from './pages/UsersPage';

// ルートルート
const rootRoute = createRootRoute({
  component: () => <Outlet />,
});

// ログイン
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
});

// レイアウト付きルート
const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'layout',
  component: Layout,
});

// ダッシュボード
const dashboardRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: DashboardPage,
});

// 勤怠
const attendanceRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/attendance',
  component: AttendancePage,
});

// 休暇申請
const leavesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/leaves',
  component: LeavesPage,
});

// シフト
const shiftsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/shifts',
  component: ShiftsPage,
});

// ユーザー管理
const usersRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/users',
  component: UsersPage,
});

// ルートツリー
export const routeTree = rootRoute.addChildren([
  loginRoute,
  layoutRoute.addChildren([
    dashboardRoute,
    attendanceRoute,
    leavesRoute,
    shiftsRoute,
    usersRoute,
  ]),
]);
