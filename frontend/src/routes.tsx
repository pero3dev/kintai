import {
  createRootRoute,
  createRoute,
  Outlet,
} from '@tanstack/react-router';
import { Layout } from './components/layout/Layout';
import { LoginPage } from './pages/LoginPage';
import { HomeDashboardPage } from './pages/HomeDashboardPage';
import { DashboardPage } from './pages/DashboardPage';
import { AttendancePage } from './pages/AttendancePage';
import { LeavesPage } from './pages/LeavesPage';
import { ShiftsPage } from './pages/ShiftsPage';
import { UsersPage } from './pages/UsersPage';
import { OvertimePage } from './pages/OvertimePage';
import { CorrectionsPage } from './pages/CorrectionsPage';
import { NotificationsPage } from './pages/NotificationsPage';
import { ProjectsPage } from './pages/ProjectsPage';
import { HolidaysPage } from './pages/HolidaysPage';
import { ExportPage } from './pages/ExportPage';
import { ApprovalFlowsPage } from './pages/ApprovalFlowsPage';

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

// ホーム
const homeRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: HomeDashboardPage,
});

// 管理者ダッシュボード
const dashboardRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/dashboard',
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

// 残業申請
const overtimeRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/overtime',
  component: OvertimePage,
});

// 勤怠修正
const correctionsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/corrections',
  component: CorrectionsPage,
});

// 通知
const notificationsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/notifications',
  component: NotificationsPage,
});

// プロジェクト
const projectsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/projects',
  component: ProjectsPage,
});

// 祝日・カレンダー
const holidaysRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/holidays',
  component: HolidaysPage,
});

// エクスポート
const exportRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/export',
  component: ExportPage,
});

// 承認フロー
const approvalFlowsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/approval-flows',
  component: ApprovalFlowsPage,
});

// ルートツリー
export const routeTree = rootRoute.addChildren([
  loginRoute,
  layoutRoute.addChildren([
    homeRoute,
    dashboardRoute,
    attendanceRoute,
    leavesRoute,
    shiftsRoute,
    usersRoute,
    overtimeRoute,
    correctionsRoute,
    notificationsRoute,
    projectsRoute,
    holidaysRoute,
    exportRoute,
    approvalFlowsRoute,
  ]),
]);
