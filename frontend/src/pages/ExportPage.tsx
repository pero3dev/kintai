import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';
import { Download } from 'lucide-react';
import { api } from '@/api/client';

export function ExportPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const [startDate, setStartDate] = useState(() => {
    const d = new Date();
    d.setMonth(d.getMonth() - 1);
    return d.toISOString().slice(0, 10);
  });
  const [endDate, setEndDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [loading, setLoading] = useState<string | null>(null);

  const isAdmin = user?.role === 'admin' || user?.role === 'manager';

  if (!isAdmin) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        {t('export.adminOnly')}
      </div>
    );
  }

  const downloadCSV = async (type: string) => {
    setLoading(type);
    try {
      let blob: Blob;
      const params = { start_date: startDate, end_date: endDate };
      switch (type) {
        case 'attendance':
          blob = await api.export.attendance(params);
          break;
        case 'leaves':
          blob = await api.export.leaves(params);
          break;
        case 'overtime':
          blob = await api.export.overtime(params);
          break;
        case 'projects':
          blob = await api.export.projects(params);
          break;
        default:
          return;
      }
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${type}_${startDate}_${endDate}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (e) {
      console.error('Export failed:', e);
    } finally {
      setLoading(null);
    }
  };

  const exportItems = [
    {
      key: 'attendance',
      title: t('export.attendance'),
      description: t('export.attendanceDesc'),
      icon: '📋',
    },
    {
      key: 'leaves',
      title: t('export.leaves'),
      description: t('export.leavesDesc'),
      icon: '🏖️',
    },
    {
      key: 'overtime',
      title: t('export.overtime'),
      description: t('export.overtimeDesc'),
      icon: '⏰',
    },
    {
      key: 'projects',
      title: t('export.projects'),
      description: t('export.projectsDesc'),
      icon: '📊',
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <Download className="h-6 w-6" />
        {t('export.title')}
      </h1>

      {/* 期間設定 */}
      <div className="glass-card rounded-2xl p-6">
        <h2 className="text-lg font-semibold mb-4">{t('export.dateRange')}</h2>
        <div className="flex items-center gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">{t('export.startDate')}</label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="px-3 py-2 glass-input rounded-xl"
            />
          </div>
          <span className="mt-6">〜</span>
          <div>
            <label className="block text-sm font-medium mb-1">{t('export.endDate')}</label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="px-3 py-2 glass-input rounded-xl"
            />
          </div>
        </div>
      </div>

      {/* エクスポート項目 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {exportItems.map((item) => (
          <div key={item.key} className="glass-card rounded-2xl p-6 flex items-start justify-between">
            <div className="flex items-start gap-4">
              <span className="text-3xl">{item.icon}</span>
              <div>
                <h3 className="font-semibold">{item.title}</h3>
                <p className="text-sm text-muted-foreground mt-1">{item.description}</p>
              </div>
            </div>
            <button
              onClick={() => downloadCSV(item.key)}
              disabled={loading === item.key}
              className="flex items-center gap-2 px-4 py-2 gradient-primary text-white rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50 whitespace-nowrap"
            >
              <Download className="h-4 w-4" />
              {loading === item.key ? t('common.downloading') : 'CSV'}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
