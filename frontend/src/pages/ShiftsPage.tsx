import { useTranslation } from 'react-i18next';
import { CalendarRange } from 'lucide-react';

export function ShiftsPage() {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <CalendarRange className="h-6 w-6" />
        {t('shifts.title')}
      </h1>

      <div className="bg-card border border-border rounded-lg p-6">
        <p className="text-muted-foreground">
          シフト管理ページ - カレンダー形式のシフト表示と編集機能を実装予定
        </p>
        {/* TODO: カレンダーコンポーネントとシフト編集UI */}
      </div>
    </div>
  );
}
