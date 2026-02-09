// アプリケーション定義
// 将来の拡張時に新しいアプリを追加できる構造

export interface AppDefinition {
  id: string;
  nameKey: string; // i18n翻訳キー
  descriptionKey: string; // i18n翻訳キー
  icon: string; // Material Symbols icon name
  color: string; // Tailwind color class
  basePath: string; // ルートベースパス
  enabled: boolean;
  comingSoon?: boolean; // 近日公開フラグ
  requiredRoles?: ('admin' | 'manager' | 'employee')[]; // 必要な権限
}

export const apps: AppDefinition[] = [
  {
    id: 'attendance',
    nameKey: 'apps.attendance.name',
    descriptionKey: 'apps.attendance.description',
    icon: 'schedule',
    color: 'bg-blue-500',
    basePath: '/',
    enabled: true,
  },
  {
    id: 'expenses',
    nameKey: 'apps.expenses.name',
    descriptionKey: 'apps.expenses.description',
    icon: 'receipt_long',
    color: 'bg-green-500',
    basePath: '/expenses',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'hr',
    nameKey: 'apps.hr.name',
    descriptionKey: 'apps.hr.description',
    icon: 'badge',
    color: 'bg-purple-500',
    basePath: '/hr',
    enabled: false,
    comingSoon: true,
    requiredRoles: ['admin', 'manager'],
  },
  {
    id: 'wiki',
    nameKey: 'apps.wiki.name',
    descriptionKey: 'apps.wiki.description',
    icon: 'menu_book',
    color: 'bg-amber-500',
    basePath: '/wiki',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'tasks',
    nameKey: 'apps.tasks.name',
    descriptionKey: 'apps.tasks.description',
    icon: 'task_alt',
    color: 'bg-cyan-500',
    basePath: '/tasks',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'booking',
    nameKey: 'apps.booking.name',
    descriptionKey: 'apps.booking.description',
    icon: 'meeting_room',
    color: 'bg-rose-500',
    basePath: '/booking',
    enabled: false,
    comingSoon: true,
  },
];

// 現在アクティブなアプリを取得
export const getActiveApp = (pathname: string): AppDefinition | undefined => {
  // 最も長いパスマッチを優先
  return apps
    .filter(app => app.enabled)
    .sort((a, b) => b.basePath.length - a.basePath.length)
    .find(app => pathname.startsWith(app.basePath));
};

// ユーザーの権限に基づいて利用可能なアプリを取得
export const getAvailableApps = (userRole?: string): AppDefinition[] => {
  return apps.filter(app => {
    if (!app.requiredRoles) return true;
    return app.requiredRoles.includes(userRole as 'admin' | 'manager' | 'employee');
  });
};
