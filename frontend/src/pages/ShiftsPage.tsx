import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { CalendarRange, ChevronLeft, ChevronRight, Plus, X } from 'lucide-react';

type ShiftType = 'morning' | 'day' | 'evening' | 'night' | 'off';

interface Shift {
  id: string;
  user_id: string;
  date: string;
  shift_type: ShiftType;
  user?: {
    id: string;
    first_name: string;
    last_name: string;
  };
}

interface User {
  id: string;
  first_name: string;
  last_name: string;
  role: string;
}

const SHIFT_TYPES: { value: ShiftType; label: string; color: string; bgColor: string }[] = [
  { value: 'morning', label: '早番', color: 'text-orange-800', bgColor: 'bg-orange-100' },
  { value: 'day', label: '日勤', color: 'text-blue-800', bgColor: 'bg-blue-100' },
  { value: 'evening', label: '遅番', color: 'text-purple-800', bgColor: 'bg-purple-100' },
  { value: 'night', label: '夜勤', color: 'text-indigo-800', bgColor: 'bg-indigo-100' },
  { value: 'off', label: '休み', color: 'text-gray-800', bgColor: 'bg-gray-100' },
];

const getShiftStyle = (type: ShiftType) => {
  return SHIFT_TYPES.find(s => s.value === type) || SHIFT_TYPES[1];
};

export function ShiftsPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const isAdmin = user?.role === 'admin' || user?.role === 'manager';

  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedCell, setSelectedCell] = useState<{ userId: string; date: string } | null>(null);
  const [selectedShiftType, setSelectedShiftType] = useState<ShiftType>('day');

  // 週の開始日と終了日を計算
  const { startDate, endDate, weekDays } = useMemo(() => {
    const start = new Date(currentDate);
    start.setDate(start.getDate() - start.getDay()); // 日曜日開始
    const end = new Date(start);
    end.setDate(end.getDate() + 6);

    const days: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const d = new Date(start);
      d.setDate(d.getDate() + i);
      days.push(d);
    }

    return {
      startDate: start.toISOString().split('T')[0],
      endDate: end.toISOString().split('T')[0],
      weekDays: days,
    };
  }, [currentDate]);

  // シフトデータ取得
  const { data: shifts = [] } = useQuery<Shift[]>({
    queryKey: ['shifts', startDate, endDate],
    queryFn: () => api.shifts.getList({ start_date: startDate, end_date: endDate }),
  });

  // ユーザー一覧取得（管理者用）
  const { data: usersData } = useQuery({
    queryKey: ['users'],
    queryFn: () => api.users.getAll(),
    enabled: isAdmin,
  });
  const users: User[] = usersData?.data || [];

  // シフト作成
  const createMutation = useMutation({
    mutationFn: (data: { user_id: string; date: string; shift_type: ShiftType }) =>
      api.shifts.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
      setSelectedCell(null);
    },
  });

  // シフト削除
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.shifts.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
    },
  });

  const navigateWeek = (direction: number) => {
    const newDate = new Date(currentDate);
    newDate.setDate(newDate.getDate() + direction * 7);
    setCurrentDate(newDate);
  };

  const getShiftForCell = (userId: string, date: string): Shift | undefined => {
    return shifts.find(s => s.user_id === userId && s.date.startsWith(date));
  };

  const handleCellClick = (userId: string, date: string) => {
    if (!isAdmin) return;
    const existing = getShiftForCell(userId, date);
    if (existing) {
      if (confirm('このシフトを削除しますか？')) {
        deleteMutation.mutate(existing.id);
      }
    } else {
      setSelectedCell({ userId, date });
    }
  };

  const handleCreateShift = () => {
    if (!selectedCell) return;
    createMutation.mutate({
      user_id: selectedCell.userId,
      date: selectedCell.date,
      shift_type: selectedShiftType,
    });
  };

  const formatDate = (date: Date) => {
    const month = date.getMonth() + 1;
    const day = date.getDate();
    const weekday = ['日', '月', '火', '水', '木', '金', '土'][date.getDay()];
    return { month, day, weekday };
  };

  // 一般ユーザーの場合は自分のシフトのみ表示
  const displayUsers = isAdmin ? users : user ? [{ id: user.id, first_name: user.first_name, last_name: user.last_name, role: user.role }] : [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <CalendarRange className="h-6 w-6" />
          {t('shifts.title')}
        </h1>

        {/* 週ナビゲーション */}
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigateWeek(-1)}
            className="p-2 hover:bg-accent rounded-lg transition-colors"
          >
            <ChevronLeft className="h-5 w-5" />
          </button>
          <span className="font-medium min-w-[200px] text-center">
            {startDate} 〜 {endDate}
          </span>
          <button
            onClick={() => navigateWeek(1)}
            className="p-2 hover:bg-accent rounded-lg transition-colors"
          >
            <ChevronRight className="h-5 w-5" />
          </button>
          <button
            onClick={() => setCurrentDate(new Date())}
            className="px-3 py-1 text-sm bg-primary text-primary-foreground rounded-lg hover:opacity-90"
          >
            今週
          </button>
        </div>
      </div>

      {/* 凡例 */}
      <div className="flex gap-4 flex-wrap">
        {SHIFT_TYPES.map(type => (
          <div key={type.value} className="flex items-center gap-2">
            <span className={`px-2 py-1 rounded text-xs ${type.bgColor} ${type.color}`}>
              {type.label}
            </span>
          </div>
        ))}
      </div>

      {/* シフトカレンダー */}
      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[800px]">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="py-3 px-4 text-left font-medium w-32">名前</th>
                {weekDays.map((date, i) => {
                  const { month, day, weekday } = formatDate(date);
                  const isToday = date.toDateString() === new Date().toDateString();
                  const isWeekend = i === 0 || i === 6;
                  return (
                    <th
                      key={i}
                      className={`py-3 px-2 text-center font-medium ${isToday ? 'bg-primary/10' : ''} ${isWeekend ? 'text-red-500' : ''}`}
                    >
                      <div className="text-xs text-muted-foreground">{month}/{day}</div>
                      <div className={isWeekend ? 'text-red-500' : ''}>{weekday}</div>
                    </th>
                  );
                })}
              </tr>
            </thead>
            <tbody>
              {displayUsers.map((u) => (
                <tr key={u.id} className="border-b border-border/50 hover:bg-accent/30">
                  <td className="py-2 px-4 font-medium">
                    {u.last_name} {u.first_name}
                  </td>
                  {weekDays.map((date, i) => {
                    const dateStr = date.toISOString().split('T')[0];
                    const shift = getShiftForCell(u.id, dateStr);
                    const isToday = date.toDateString() === new Date().toDateString();
                    const style = shift ? getShiftStyle(shift.shift_type) : null;

                    return (
                      <td
                        key={i}
                        onClick={() => handleCellClick(u.id, dateStr)}
                        className={`py-2 px-2 text-center ${isToday ? 'bg-primary/5' : ''} ${isAdmin ? 'cursor-pointer hover:bg-accent/50' : ''}`}
                      >
                        {shift && style && (
                          <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${style.bgColor} ${style.color}`}>
                            {style.label}
                          </span>
                        )}
                      </td>
                    );
                  })}
                </tr>
              ))}
              {displayUsers.length === 0 && (
                <tr>
                  <td colSpan={8} className="py-8 text-center text-muted-foreground">
                    {isAdmin ? 'ユーザーがいません' : 'シフトデータがありません'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* シフト作成モーダル */}
      {selectedCell && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">シフト登録</h2>
              <button onClick={() => setSelectedCell(null)} className="p-1 hover:bg-accent rounded">
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">日付</label>
                <div className="text-muted-foreground">{selectedCell.date}</div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">シフト種別</label>
                <div className="grid grid-cols-5 gap-2">
                  {SHIFT_TYPES.map(type => (
                    <button
                      key={type.value}
                      onClick={() => setSelectedShiftType(type.value)}
                      className={`px-3 py-2 rounded text-sm font-medium transition-all ${selectedShiftType === type.value
                          ? `${type.bgColor} ${type.color} ring-2 ring-primary`
                          : 'bg-muted hover:bg-accent'
                        }`}
                    >
                      {type.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex gap-2 pt-4">
                <button
                  onClick={() => setSelectedCell(null)}
                  className="flex-1 px-4 py-2 border border-border rounded-lg hover:bg-accent"
                >
                  キャンセル
                </button>
                <button
                  onClick={handleCreateShift}
                  disabled={createMutation.isPending}
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  <Plus className="h-4 w-4" />
                  登録
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
