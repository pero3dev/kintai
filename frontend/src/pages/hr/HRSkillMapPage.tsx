import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const LEVEL_COLORS = ['bg-gray-500/30', 'bg-red-500/30', 'bg-amber-500/30', 'bg-blue-500/30', 'bg-green-500/30'];
const LEVEL_WIDTHS = [0, 25, 50, 75, 100];

export function HRSkillMapPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [department, setDepartment] = useState('');
  const [viewMode, setViewMode] = useState<'map' | 'gap'>('map');
  const [showAddSkill, setShowAddSkill] = useState(false);
  const [selectedEmployee, setSelectedEmployee] = useState<string>('');
  const [newSkill, setNewSkill] = useState({ skill_name: '', category: 'technical', level: 3 });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-skill-map', department],
    queryFn: () => api.hr.getSkillMap({ department }),
  });
  const { data: gapData } = useQuery({
    queryKey: ['hr-skill-gap', department],
    queryFn: () => api.hr.getSkillGapAnalysis({ department }),
    enabled: viewMode === 'gap',
  });

  const addSkillMutation = useMutation({
    mutationFn: ({ empId, skill }: { empId: string; skill: typeof newSkill }) =>
      api.hr.addEmployeeSkill(empId, skill),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-skill-map'] });
      setShowAddSkill(false);
      setNewSkill({ skill_name: '', category: 'technical', level: 3 });
    },
  });

  const employees: Record<string, unknown>[] = data?.data || data || [];
  const gapAnalysis: Record<string, unknown>[] = gapData?.data || gapData || [];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.skillMap.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.skillMap.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <input value={department} onChange={e => setDepartment(e.target.value)}
            placeholder={t('hr.skillMap.department')} className="px-4 py-2.5 glass-input rounded-xl text-sm w-32" />
          <div className="flex glass-subtle rounded-xl overflow-hidden">
            <button onClick={() => setViewMode('map')}
              className={`px-3 py-2 text-sm flex items-center gap-1 ${viewMode === 'map' ? 'bg-white/10' : ''}`}>
              <MaterialIcon name="grid_view" className="text-sm" />
              {t('hr.skillMap.skillMap')}
            </button>
            <button onClick={() => setViewMode('gap')}
              className={`px-3 py-2 text-sm flex items-center gap-1 ${viewMode === 'gap' ? 'bg-white/10' : ''}`}>
              <MaterialIcon name="analytics" className="text-sm" />
              {t('hr.skillMap.gapAnalysis')}
            </button>
          </div>
        </div>
      </div>

      {/* スキル追加モーダル */}
      {showAddSkill && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setShowAddSkill(false)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-md relative z-10 animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">{t('hr.skillMap.addSkill')}</h3>
            <div className="space-y-3">
              <input value={newSkill.skill_name} onChange={e => setNewSkill(p => ({ ...p, skill_name: e.target.value }))}
                placeholder={t('hr.skillMap.skillName')} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <select value={newSkill.category} onChange={e => setNewSkill(p => ({ ...p, category: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                {['technical', 'soft', 'management', 'domain'].map(c => (
                  <option key={c} value={c}>{t(`hr.skillMap.categories.${c}`)}</option>
                ))}
              </select>
              <div>
                <label className="text-xs text-muted-foreground block mb-1">{t('hr.skillMap.currentLevel')}: {t(`hr.skillMap.levels.${newSkill.level}`)}</label>
                <input type="range" min={1} max={5} value={newSkill.level}
                  onChange={e => setNewSkill(p => ({ ...p, level: Number(e.target.value) }))}
                  className="w-full" />
                <div className="flex justify-between text-[10px] text-muted-foreground mt-1">
                  {[1,2,3,4,5].map(l => <span key={l}>{t(`hr.skillMap.levels.${l}`)}</span>)}
                </div>
              </div>
            </div>
            <div className="flex gap-2 mt-4 justify-end">
              <button onClick={() => setShowAddSkill(false)} className="px-4 py-2 glass-subtle rounded-xl text-sm">{t('common.cancel')}</button>
              <button onClick={() => selectedEmployee && addSkillMutation.mutate({ empId: selectedEmployee, skill: newSkill })}
                className="gradient-primary px-4 py-2 rounded-xl text-sm font-medium">{t('common.save')}</button>
            </div>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : viewMode === 'gap' ? (
        /* ギャップ分析 */
        gapAnalysis.length === 0 ? (
          <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
            <MaterialIcon name="analytics" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('common.noData')}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {gapAnalysis.map((gap, i) => (
              <div key={i} className="glass-card rounded-2xl p-4 sm:p-5">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="font-semibold text-sm">{String(gap.skill_name || '')}</h3>
                  <span className="text-[10px] glass-subtle px-2 py-0.5 rounded">{t(`hr.skillMap.categories.${String(gap.category || 'technical')}`)}</span>
                </div>
                <div className="grid grid-cols-2 gap-4 text-center mb-3">
                  <div>
                    <p className="text-xs text-muted-foreground">{t('hr.skillMap.currentLevel')}</p>
                    <p className="text-xl font-bold text-amber-400">{String(gap.current_avg || 0)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">{t('hr.skillMap.requiredLevel')}</p>
                    <p className="text-xl font-bold text-green-400">{String(gap.required_level || 0)}</p>
                  </div>
                </div>
                <div className="relative h-2 rounded-full bg-white/10">
                  <div className="absolute top-0 left-0 h-full rounded-full bg-amber-400/60 transition-all"
                    style={{ width: `${(Number(gap.current_avg || 0) / 5) * 100}%` }} />
                  <div className="absolute top-0 h-full w-0.5 bg-green-400"
                    style={{ left: `${(Number(gap.required_level || 0) / 5) * 100}%` }} />
                </div>
                <p className="text-[10px] text-muted-foreground mt-2">
                  {t('hr.skillMap.gapAnalysis')}: {Number(gap.gap || 0) > 0 ? `+${gap.gap}` : String(gap.gap || '0')}
                </p>
              </div>
            ))}
          </div>
        )
      ) : (
        /* スキルマップ */
        employees.length === 0 ? (
          <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
            <MaterialIcon name="psychology" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('common.noData')}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {employees.map((emp) => {
              const skills: Record<string, unknown>[] = Array.isArray(emp.skills) ? emp.skills as Record<string, unknown>[] : [];
              return (
                <div key={String(emp.id)} className="glass-card rounded-2xl p-4 sm:p-5">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-3">
                      <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400 font-bold">
                        {String(emp.name || '?').substring(0, 1)}
                      </div>
                      <div>
                        <p className="font-semibold text-sm">{String(emp.name)}</p>
                        <p className="text-[10px] text-muted-foreground">{String(emp.position || '')} · {String(emp.department || '')}</p>
                      </div>
                    </div>
                    <button onClick={() => { setSelectedEmployee(String(emp.id)); setShowAddSkill(true); }}
                      className="p-2 glass-subtle rounded-xl hover:bg-white/10 text-xs flex items-center gap-1">
                      <MaterialIcon name="add" className="text-xs" />
                      {t('hr.skillMap.addSkill')}
                    </button>
                  </div>
                  {skills.length === 0 ? (
                    <p className="text-xs text-muted-foreground text-center py-4">{t('common.noData')}</p>
                  ) : (
                    <div className="space-y-2">
                      {skills.map((skill, si) => {
                        const level = Math.min(Math.max(Number(skill.level || 0), 0), 5);
                        return (
                          <div key={si} className="flex items-center gap-3">
                            <span className={`text-[10px] px-2 py-0.5 rounded glass-subtle shrink-0 w-16 text-center`}>
                              {t(`hr.skillMap.categories.${String(skill.category || 'technical')}`)}
                            </span>
                            <span className="text-sm w-28 truncate shrink-0">{String(skill.skill_name || '')}</span>
                            <div className="flex-1 h-2 rounded-full bg-white/10 relative">
                              <div className={`h-full rounded-full ${LEVEL_COLORS[level] || 'bg-blue-500/30'} transition-all`}
                                style={{ width: `${LEVEL_WIDTHS[level] || 0}%` }} />
                            </div>
                            <span className="text-[10px] text-muted-foreground w-16 text-right shrink-0">
                              {t(`hr.skillMap.levels.${level}`)}
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )
      )}
    </div>
  );
}
