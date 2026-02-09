import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';
import { Calendar, Plus, ChevronLeft, ChevronRight } from 'lucide-react';

const createHolidaySchema = (t: (key: string) => string) => z.object({
  date: z.string().min(1, t('holidays.validation.dateRequired')),
  name: z.string().min(1, t('holidays.validation.nameRequired')),
  holiday_type: z.enum(['national', 'company', 'optional']),
  is_recurring: z.boolean().optional(),
});

type HolidayForm = z.infer<ReturnType<typeof createHolidaySchema>>;

export function HolidaysPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [currentYear, setCurrentYear] = useState(new Date().getFullYear());
  const [currentMonth, setCurrentMonth] = useState(new Date().getMonth() + 1);
  const isAdmin = user?.role === 'admin' || user?.role === 'manager';
  const holidaySchema = createHolidaySchema(t);

  const { data: holidays } = useQuery({
    queryKey: ['holidays', currentYear],
    queryFn: () => api.holidays.getByYear({ year: String(currentYear) }),
  });

  const { data: calendar } = useQuery({
    queryKey: ['holidays', 'calendar', currentYear, currentMonth],
    queryFn: () => api.holidays.getCalendar({ year: String(currentYear), month: String(currentMonth) }),
  });

  const { data: workingDays } = useQuery({
    queryKey: ['holidays', 'working-days', currentYear, currentMonth],
    queryFn: () => api.holidays.getWorkingDays({
      start_date: `${currentYear}-${String(currentMonth).padStart(2, '0')}-01`,
      end_date: `${currentYear}-${String(currentMonth).padStart(2, '0')}-${new Date(currentYear, currentMonth, 0).getDate()}`,
    }),
  });

  const { register, handleSubmit, reset, formState: { errors } } = useForm<HolidayForm>({
    resolver: zodResolver(holidaySchema),
    defaultValues: { holiday_type: 'national' },
  });

  const createMutation = useMutation({
    mutationFn: (data: HolidayForm) => api.holidays.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holidays'] });
      setShowForm(false);
      reset();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.holidays.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['holidays'] });
    },
  });

  const prevMonth = () => {
    if (currentMonth === 1) {
      setCurrentMonth(12);
      setCurrentYear(currentYear - 1);
    } else {
      setCurrentMonth(currentMonth - 1);
    }
  };

  const nextMonth = () => {
    if (currentMonth === 12) {
      setCurrentMonth(1);
      setCurrentYear(currentYear + 1);
    } else {
      setCurrentMonth(currentMonth + 1);
    }
  };

  const typeBadge = (type: string) => {
    const styles: Record<string, string> = {
      national: 'bg-red-500/20 text-red-400 border border-red-500/30',
      company: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
      optional: 'bg-purple-500/20 text-purple-400 border border-purple-500/30',
    };
    const labels: Record<string, string> = {
      national: t('holidays.types.national'),
      company: t('holidays.types.company'),
      optional: t('holidays.types.optional'),
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs ${styles[type] || ''}`}>
        {labels[type] || type}
      </span>
    );
  };

  // カレンダーグリッド生成
  const calendarGrid = () => {
    if (!calendar) return [];
    const firstDay = new Date(currentYear, currentMonth - 1, 1).getDay();
    const cells: (Record<string, unknown> | null)[] = [];
    for (let i = 0; i < firstDay; i++) cells.push(null);
    for (const day of calendar as Record<string, unknown>[]) {
      cells.push(day);
    }
    return cells;
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Calendar className="h-6 w-6" />
          {t('holidays.title')}
        </h1>
        {isAdmin && (
          <button
            onClick={() => setShowForm(!showForm)}
            className="flex items-center gap-2 px-4 py-2 gradient-primary text-white rounded-xl hover:shadow-glow-md transition-all"
          >
            <Plus className="h-4 w-4" />
            {t('holidays.add')}
          </button>
        )}
      </div>

      {/* 追加フォーム */}
      {showForm && (
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4">{t('holidays.add')}</h2>
          <form onSubmit={handleSubmit((data) => createMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('holidays.date')}</label>
                <input type="date" {...register('date')} className="w-full px-3 py-2 glass-input rounded-xl" />
                {errors.date && <p className="text-sm text-red-400 mt-1">{errors.date.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('holidays.name')}</label>
                <input type="text" {...register('name')} className="w-full px-3 py-2 glass-input rounded-xl" />
                {errors.name && <p className="text-sm text-red-400 mt-1">{errors.name.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('holidays.type')}</label>
                <select {...register('holiday_type')} className="w-full px-3 py-2 glass-input rounded-xl">
                  <option value="national">{t('holidays.types.national')}</option>
                  <option value="company">{t('holidays.types.company')}</option>
                  <option value="optional">{t('holidays.types.optional')}</option>
                </select>
              </div>
              <div className="flex items-center gap-2 pt-6">
                <input type="checkbox" {...register('is_recurring')} id="is_recurring" className="rounded" />
                <label htmlFor="is_recurring" className="text-sm">{t('holidays.recurring')}</label>
              </div>
            </div>
            <div className="flex gap-2">
              <button type="submit" className="px-4 py-2 gradient-primary text-white rounded-xl hover:shadow-glow-md transition-all">
                {t('common.create')}
              </button>
              <button type="button" onClick={() => setShowForm(false)} className="px-4 py-2 glass-input rounded-xl hover:bg-white/10 transition-all">
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Calendar */}
        <div className="lg:col-span-2 glass-card rounded-2xl p-6">
          <div className="flex items-center justify-between mb-6">
            <button onClick={prevMonth} className="p-2 hover:bg-white/10 transition-colors rounded-md">
              <ChevronLeft className="h-5 w-5" />
            </button>
            <h2 className="text-lg font-semibold">
              {t('holidays.yearMonth', { year: currentYear, month: currentMonth })}
            </h2>
            <button onClick={nextMonth} className="p-2 hover:bg-white/10 transition-colors rounded-md">
              <ChevronRight className="h-5 w-5" />
            </button>
          </div>

          {/* Weekday Header */}
          <div className="grid grid-cols-7 gap-1 mb-2">
            {['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'].map((dow, idx) => (
              <div key={dow} className={`text-center text-xs font-medium py-2 ${idx === 0 ? 'text-red-500' : idx === 6 ? 'text-blue-500' : 'text-muted-foreground'}`}>
                {t(`common.weekdays.${dow}`)}
              </div>
            ))}
          </div>

          {/* カレンダーグリッド */}
          <div className="grid grid-cols-7 gap-1">
            {calendarGrid().map((cell, i) => {
              if (!cell) return <div key={`empty-${i}`} className="p-2 h-16" />;
              const date = (cell.date as string)?.slice(8, 10);
              const isHoliday = cell.is_holiday as boolean;
              const isWeekend = cell.is_weekend as boolean;
              return (
                <div
                  key={cell.date as string}
                  className={`p-2 h-16 rounded text-sm ${
                    isHoliday
                      ? 'bg-red-500/10 border border-red-500/20'
                      : isWeekend
                        ? 'bg-muted/50'
                        : 'hover:bg-white/5 transition-colors'
                  }`}
                >
                  <span className={`text-xs font-medium ${isHoliday || isWeekend ? 'text-red-400' : ''}`}>
                    {parseInt(date)}
                  </span>
                  {isHoliday && (
                    <p className="text-[10px] text-red-600 truncate mt-0.5">
                      {cell.holiday_name as string}
                    </p>
                  )}
                </div>
              );
            })}
          </div>

          {/* Working Days Summary */}
          {workingDays && (
            <div className="grid grid-cols-4 gap-4 mt-4 pt-4 border-t border-border">
              <div className="text-center">
                <p className="text-2xl font-bold">{workingDays.working_days}</p>
                <p className="text-xs text-muted-foreground">{t('holidays.workingDays')}</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold">{workingDays.holidays}</p>
                <p className="text-xs text-muted-foreground">{t('holidays.holidayCount')}</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold">{workingDays.weekends}</p>
                <p className="text-xs text-muted-foreground">{t('holidays.weekends')}</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold">{workingDays.total_days}</p>
                <p className="text-xs text-muted-foreground">{t('holidays.totalDays')}</p>
              </div>
            </div>
          )}
        </div>

        {/* Holiday List */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4">{t('holidays.yearHolidays', { year: currentYear })}</h2>
          <div className="space-y-3 max-h-[600px] overflow-y-auto">
            {(holidays as Record<string, unknown>[] | undefined)?.map((h: Record<string, unknown>) => (
              <div key={h.id as string} className="flex items-center justify-between p-3 glass-subtle rounded-xl">
                <div>
                  <p className="font-medium text-sm">{h.name as string}</p>
                  <p className="text-xs text-muted-foreground">{(h.date as string)?.slice(0, 10)}</p>
                  {typeBadge(h.holiday_type as string)}
                </div>
                {isAdmin && (
                  <button
                    onClick={() => deleteMutation.mutate(h.id as string)}
                    className="text-destructive hover:underline text-xs"
                  >
                    {t('common.delete')}
                  </button>
                )}
              </div>
            ))}
            {(!holidays || (holidays as unknown[]).length === 0) && (
              <p className="text-center text-muted-foreground text-sm py-4">{t('common.noData')}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
