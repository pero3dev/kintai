import { useState, useRef, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
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
  const triggerRef = useRef<HTMLButtonElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [dropdownPos, setDropdownPos] = useState({ top: 0, left: 0 });

  const activeApp = getActiveApp(location.pathname);
  const availableApps = getAvailableApps(user?.role);

  // ドロップダウンの位置を計算
  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    if (collapsed) {
      // サイドバー折り畳み時: ボタンの右側に表示
      setDropdownPos({ top: rect.top, left: rect.right + 8 });
    } else {
      // 通常時: ボタンの下に表示
      setDropdownPos({ top: rect.bottom + 8, left: rect.left });
    }
  }, [collapsed]);

  // 外側クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current && !dropdownRef.current.contains(event.target as Node) &&
        triggerRef.current && !triggerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // 開いたとき位置を更新 + スクロール/リサイズ追従
  useEffect(() => {
    if (!isOpen) return;
    updatePosition();
    window.addEventListener('resize', updatePosition);
    window.addEventListener('scroll', updatePosition, true);
    return () => {
      window.removeEventListener('resize', updatePosition);
      window.removeEventListener('scroll', updatePosition, true);
    };
  }, [isOpen, updatePosition]);

  const handleAppSelect = (app: AppDefinition) => {
    if (app.enabled && !app.comingSoon) {
      navigate({ to: app.basePath as '/' });
      setIsOpen(false);
    }
  };

  return (
    <div className="relative">
      {/* トリガーボタン */}
      <button
        ref={triggerRef}
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-2 w-full rounded-xl transition-all duration-200 nav-item-hover ${
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

      {/* ドロップダウンメニュー（Portal で body 直下に描画） */}
      {isOpen && createPortal(
        <div
          ref={dropdownRef}
          className="fixed z-[9999] glass rounded-2xl shadow-2xl overflow-hidden animate-scale-in"
          style={{ top: dropdownPos.top, left: dropdownPos.left, minWidth: '280px' }}
        >
          {/* ヘッダー */}
          <div className="px-4 py-3 border-b border-white/5">
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
                    ? 'hover:bg-white/5 cursor-pointer'
                    : 'opacity-40 cursor-not-allowed'
                } ${activeApp?.id === app.id ? 'glass-subtle ring-1 ring-indigo-400/30' : ''}`}
              >
                <div className={`size-12 rounded-xl flex items-center justify-center text-white ${app.color}`}>
                  <MaterialIcon name={app.icon} className="text-2xl" />
                </div>
                <div className="text-center">
                  <p className="text-sm font-medium">{t(app.nameKey)}</p>
                  {app.comingSoon && (
                    <span className="inline-block mt-1 px-2 py-0.5 bg-amber-400/10 text-amber-400 border border-amber-400/20 text-[10px] font-bold rounded-full">
                      {t('appSwitcher.comingSoon')}
                    </span>
                  )}
                </div>
              </button>
            ))}
          </div>

          {/* フッター */}
          <div className="px-4 py-3 border-t border-white/5">
            <p className="text-xs text-muted-foreground text-center">
              {t('appSwitcher.moreApps')}
            </p>
          </div>
        </div>,
        document.body
      )}
    </div>
  );
}
