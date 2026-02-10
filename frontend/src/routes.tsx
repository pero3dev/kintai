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
import { ExpenseDashboardPage } from './pages/expenses/ExpenseDashboardPage';
import { ExpenseNewPage } from './pages/expenses/ExpenseNewPage';
import { ExpenseHistoryPage } from './pages/expenses/ExpenseHistoryPage';
import { ExpenseApprovePage } from './pages/expenses/ExpenseApprovePage';
import { ExpenseDetailPage } from './pages/expenses/ExpenseDetailPage';
import { ExpenseReportPage } from './pages/expenses/ExpenseReportPage';
import { ExpenseTemplatesPage } from './pages/expenses/ExpenseTemplatesPage';
import { ExpensePolicyPage } from './pages/expenses/ExpensePolicyPage';
import { ExpenseNotificationsPage } from './pages/expenses/ExpenseNotificationsPage';
import { ExpenseAdvancedApprovePage } from './pages/expenses/ExpenseAdvancedApprovePage';
import { HRDashboardPage } from './pages/hr/HRDashboardPage';
import { HREmployeesPage } from './pages/hr/HREmployeesPage';
import { HREmployeeDetailPage } from './pages/hr/HREmployeeDetailPage';
import { HRDepartmentsPage } from './pages/hr/HRDepartmentsPage';
import { HREvaluationsPage } from './pages/hr/HREvaluationsPage';
import { HRGoalsPage } from './pages/hr/HRGoalsPage';
import { HRTrainingPage } from './pages/hr/HRTrainingPage';
import { HRRecruitmentPage } from './pages/hr/HRRecruitmentPage';
import { HRDocumentsPage } from './pages/hr/HRDocumentsPage';
import { HRAnnouncementsPage } from './pages/hr/HRAnnouncementsPage';
import { HRAttendanceIntegrationPage } from './pages/hr/HRAttendanceIntegrationPage';
import { HROrgChartPage } from './pages/hr/HROrgChartPage';
import { HROneOnOnePage } from './pages/hr/HROneOnOnePage';
import { HRSkillMapPage } from './pages/hr/HRSkillMapPage';
import { HRSalarySimulatorPage } from './pages/hr/HRSalarySimulatorPage';
import { HROnboardingPage } from './pages/hr/HROnboardingPage';
import { HROffboardingPage } from './pages/hr/HROffboardingPage';
import { HRSurveyPage } from './pages/hr/HRSurveyPage';
import { WikiPage } from './pages/wiki/WikiPage';

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

// 経費精算
const expenseDashboardRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses',
  component: ExpenseDashboardPage,
});

const expenseNewRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/new',
  component: ExpenseNewPage,
});

const expenseHistoryRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/history',
  component: ExpenseHistoryPage,
});

const expenseApproveRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/approve',
  component: ExpenseApprovePage,
});

const expenseDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/$expenseId',
  component: ExpenseDetailPage,
});

const expenseReportRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/report',
  component: ExpenseReportPage,
});

const expenseTemplatesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/templates',
  component: ExpenseTemplatesPage,
});

const expensePolicyRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/policy',
  component: ExpensePolicyPage,
});

const expenseNotificationsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/notifications',
  component: ExpenseNotificationsPage,
});

const expenseAdvancedApproveRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/advanced-approve',
  component: ExpenseAdvancedApprovePage,
});

// 人事管理
const hrDashboardRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr',
  component: HRDashboardPage,
});

const hrEmployeesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/employees',
  component: HREmployeesPage,
});

const hrEmployeeDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/employees/$employeeId',
  component: HREmployeeDetailPage,
});

const hrDepartmentsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/departments',
  component: HRDepartmentsPage,
});

const hrEvaluationsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/evaluations',
  component: HREvaluationsPage,
});

const hrGoalsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/goals',
  component: HRGoalsPage,
});

const hrTrainingRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/training',
  component: HRTrainingPage,
});

const hrRecruitmentRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/recruitment',
  component: HRRecruitmentPage,
});

const hrDocumentsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/documents',
  component: HRDocumentsPage,
});

const hrAnnouncementsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/announcements',
  component: HRAnnouncementsPage,
});

const hrAttendanceIntegrationRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/attendance-integration',
  component: HRAttendanceIntegrationPage,
});

const hrOrgChartRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/org-chart',
  component: HROrgChartPage,
});

const hrOneOnOneRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/one-on-one',
  component: HROneOnOnePage,
});

const hrSkillMapRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/skill-map',
  component: HRSkillMapPage,
});

const hrSalarySimulatorRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/salary',
  component: HRSalarySimulatorPage,
});

const hrOnboardingRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/onboarding',
  component: HROnboardingPage,
});

const hrOffboardingRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/offboarding',
  component: HROffboardingPage,
});

const hrSurveyRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/hr/survey',
  component: HRSurveyPage,
});

// 社内Wiki
const wikiHomeRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/wiki',
  component: WikiPage,
});

const wikiArchitectureRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/wiki/architecture',
  component: WikiPage,
});

const wikiBackendRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/wiki/backend',
  component: WikiPage,
});

const wikiFrontendRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/wiki/frontend',
  component: WikiPage,
});

const wikiInfrastructureRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/wiki/infrastructure',
  component: WikiPage,
});

const wikiTestingRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/wiki/testing',
  component: WikiPage,
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
    expenseDashboardRoute,
    expenseNewRoute,
    expenseHistoryRoute,
    expenseApproveRoute,
    expenseDetailRoute,
    expenseReportRoute,
    expenseTemplatesRoute,
    expensePolicyRoute,
    expenseNotificationsRoute,
    expenseAdvancedApproveRoute,
    hrDashboardRoute,
    hrEmployeesRoute,
    hrEmployeeDetailRoute,
    hrDepartmentsRoute,
    hrEvaluationsRoute,
    hrGoalsRoute,
    hrTrainingRoute,
    hrRecruitmentRoute,
    hrDocumentsRoute,
    hrAnnouncementsRoute,
    hrAttendanceIntegrationRoute,
    hrOrgChartRoute,
    hrOneOnOneRoute,
    hrSkillMapRoute,
    hrSalarySimulatorRoute,
    hrOnboardingRoute,
    hrOffboardingRoute,
    hrSurveyRoute,
    wikiHomeRoute,
    wikiArchitectureRoute,
    wikiBackendRoute,
    wikiFrontendRoute,
    wikiInfrastructureRoute,
    wikiTestingRoute,
  ]),
]);
