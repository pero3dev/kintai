import { useState, useRef, useEffect } from 'react';
import { useNavigate, useLocation } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { getActiveApp, getAvailableApps, AppDefinition } from '@/config/apps';
import { useAuthStore } from '@/stores/authStore';

// Material Symbols icon component
function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

interface AppSwitcherProps {
  collapsed?: boolean;
}

export function AppSwitcher({ collapsed = false }: AppSwitcherProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuthStore();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const activeApp = getActiveApp(location.pathname);
  const availableApps = getAvailableApps(user?.role);

  // 外側クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleAppSelect = (app: AppDefinition) => {
    if (app.enabled && !app.comingSoon) {
      navigate({ to: app.basePath as '/' });
      setIsOpen(false);
    }
  };

  return (
    <div ref={dropdownRef} className="relative">
      {/* トリガーボタン */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-2 w-full rounded-lg transition-all duration-200 hover:bg-primary/10 ${
          collapsed ? 'justify-center p-2' : 'px-3 py-2'
        }`}
        title={collapsed ? t('appSwitcher.title') : undefined}
      >
        <div className={`size-8 rounded-lg flex items-center justify-center text-white ${activeApp?.color || 'bg-primary'}`}>
          <MaterialIcon name="apps" className="text-lg" />
        </div>
        {!collapsed && (
          <>
            <div className="flex-1 text-left">
              <p className="text-xs text-muted-foreground">{t('appSwitcher.currentApp')}</p>
              <p className="text-sm font-semibold truncate">
                {activeApp ? t(activeApp.nameKey) : t('apps.attendance.name')}
              </p>
            </div>
            <MaterialIcon 
              name={isOpen ? 'expand_less' : 'expand_more'} 
              className="text-muted-foreground" 
            />
          </>
        )}
      </button>

      {/* ドロップダウンメニュー */}
      {isOpen && (
        <div className={`absolute z-50 mt-2 bg-card border border-border rounded-xl shadow-2xl overflow-hidden ${
          collapsed ? 'left-full ml-2 top-0' : 'left-0 right-0'
        }`} style={{ minWidth: '280px' }}>
          {/* ヘッダー */}
          <div className="px-4 py-3 border-b border-border bg-muted/30">
            <p className="font-semibold text-sm">{t('appSwitcher.title')}</p>
            <p className="text-xs text-muted-foreground">{t('appSwitcher.subtitle')}</p>
          </div>

          {/* アプリグリッド */}
          <div className="p-3 grid grid-cols-2 gap-2">
            {availableApps.map((app) => (
              <button
                key={app.id}
                onClick={() => handleAppSelect(app)}
                disabled={!app.enabled || app.comingSoon}
                className={`flex flex-col items-center gap-2 p-4 rounded-xl transition-all duration-200 ${
                  app.enabled && !app.comingSoon
                    ? 'hover:bg-primary/10 cursor-pointer'
                    : 'opacity-50 cursor-not-allowed'
                } ${activeApp?.id === app.id ? 'bg-primary/10 ring-2 ring-primary' : ''}`}
              >
                <div className={`size-12 rounded-xl flex items-center justify-center text-white ${app.color}`}>
                  <MaterialIcon name={app.icon} className="text-2xl" />
                </div>
                <div className="text-center">
                  <p className="text-sm font-medium">{t(app.nameKey)}</p>
                  {app.comingSoon && (
                    <span className="inline-block mt-1 px-2 py-0.5 bg-amber-500/20 text-amber-500 text-[10px] font-bold rounded-full">
                      {t('appSwitcher.comingSoon')}
                    </span>
                  )}
                </div>
              </button>
            ))}
          </div>

          {/* フッター */}
          <div className="px-4 py-3 border-t border-border bg-muted/30">
            <p className="text-xs text-muted-foreground text-center">
              {t('appSwitcher.moreApps')}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
