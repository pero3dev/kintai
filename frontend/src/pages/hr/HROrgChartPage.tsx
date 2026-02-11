import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

interface OrgNode {
  id: string;
  name: string;
  position: string;
  department: string;
  avatar?: string;
  children?: OrgNode[];
  employee_count?: number;
}

function OrgTreeNode({ node, depth = 0, isFlat = false }: { node: OrgNode; depth?: number; isFlat?: boolean }) {
  const [expanded, setExpanded] = useState(depth < 2);
  const hasChildren = node.children && node.children.length > 0;

  if (isFlat) {
    return (
      <>
        <div className="glass-subtle rounded-xl p-3 flex items-center gap-3 hover:bg-white/10 transition-all"
          style={{ marginLeft: `${depth * 20}px` }}>
          <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400 font-bold text-sm shrink-0">
            {node.name.substring(0, 1)}
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-semibold text-sm truncate">{node.name}</p>
            <p className="text-xs text-muted-foreground truncate">{node.position}</p>
          </div>
          <span className="text-[10px] text-muted-foreground px-2 py-0.5 glass-subtle rounded">{node.department}</span>
          {node.employee_count != null && (
            <span className="text-[10px] text-muted-foreground">
              <MaterialIcon name="group" className="text-xs align-middle" /> {node.employee_count}
            </span>
          )}
        </div>
        {expanded && hasChildren && node.children!.map(child => (
          <OrgTreeNode key={child.id} node={child} depth={depth + 1} isFlat />
        ))}
      </>
    );
  }

  return (
    <div className="flex flex-col items-center">
      <div className="glass-card rounded-2xl p-3 sm:p-4 text-center min-w-[140px] max-w-[200px] relative cursor-pointer hover:bg-white/10 transition-all"
        onClick={() => hasChildren && setExpanded(!expanded)}>
        <div className="size-12 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400 font-bold mx-auto mb-2">
          {node.name.substring(0, 1)}
        </div>
        <p className="font-semibold text-sm truncate">{node.name}</p>
        <p className="text-[10px] text-muted-foreground truncate">{node.position}</p>
        <p className="text-[10px] text-indigo-400 truncate">{node.department}</p>
        {hasChildren && (
          <MaterialIcon name={expanded ? 'expand_less' : 'expand_more'}
            className="absolute -bottom-2 left-1/2 -translate-x-1/2 text-sm glass-subtle rounded-full size-5 flex items-center justify-center" />
        )}
      </div>
      {expanded && hasChildren && (
        <div className="mt-6 relative">
          <div className="absolute top-0 left-1/2 w-px h-4 bg-white/20 -translate-y-2" />
          {node.children!.length > 1 && (
            <div className="absolute top-4 left-0 right-0 h-px bg-white/20 -translate-y-2" />
          )}
          <div className="flex gap-4 pt-4">
            {node.children!.map(child => (
              <div key={child.id} className="flex flex-col items-center relative">
                <div className="absolute top-0 left-1/2 w-px h-4 bg-white/20 -translate-y-4" />
                <OrgTreeNode node={child} depth={depth + 1} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export function HROrgChartPage() {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<'tree' | 'flat'>('tree');
  const [search, setSearch] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['hr-org-chart'],
    queryFn: () => api.hr.getOrgChart(),
  });

  // バックエンドは部署の配列を返すため、ツリー構造に変換
  const orgData: OrgNode = (() => {
    const raw = data?.data || data;
    if (!raw) return { id: '0', name: '', position: '', department: '', children: [] };

    // 単一オブジェクト（既にツリー形式）の場合
    if (!Array.isArray(raw)) {
      return raw as OrgNode;
    }

    // 部署配列 → ツリー変換
    type DeptNode = {
      id: string;
      name?: string;
      parent_id?: string | null;
      manager_id?: string | null;
      employees?: { id: string; name?: string; position?: string }[];
      employee_count?: number;
    };
    const depts = raw as DeptNode[];

    const deptMap = new Map<string, OrgNode>();
    for (const dept of depts) {
      const empChildren: OrgNode[] = (dept.employees || []).map((e) => ({
        id: String(e.id),
        name: e.name || '—',
        position: e.position || '',
        department: dept.name || '',
      }));
      deptMap.set(String(dept.id), {
        id: String(dept.id),
        name: dept.name || '—',
        position: '',
        department: '',
        employee_count: empChildren.length,
        children: empChildren,
      });
    }

    // 親子関係を構築
    const roots: OrgNode[] = [];
    for (const dept of depts) {
      const node = deptMap.get(String(dept.id))!;
      if (dept.parent_id) {
        const parent = deptMap.get(String(dept.parent_id));
        if (parent) {
          parent.children = [...(parent.children || []), node];
          continue;
        }
      }
      roots.push(node);
    }

    if (roots.length === 1) return roots[0];
    return {
      id: 'root',
      name: t('hr.orgChart.title'),
      position: '',
      department: '',
      children: roots,
    };
  })();

  function filterNodes(node: OrgNode, query: string): OrgNode | null {
    if (!query) return node;
    const q = query.toLowerCase();
    const matches = node.name.toLowerCase().includes(q) ||
      node.position.toLowerCase().includes(q) ||
      node.department.toLowerCase().includes(q);
    const filteredChildren = (node.children || []).map(c => filterNodes(c, query)).filter(Boolean) as OrgNode[];
    if (matches || filteredChildren.length > 0) {
      return { ...node, children: filteredChildren };
    }
    return null;
  }

  const filteredData = filterNodes(orgData, search);

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.orgChart.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.orgChart.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <input value={search} onChange={(e) => setSearch(e.target.value)}
            placeholder={t('hr.orgChart.searchPlaceholder')}
            className="px-4 py-2.5 glass-input rounded-xl text-sm w-40" />
          <div className="flex glass-subtle rounded-xl overflow-hidden">
            <button onClick={() => setViewMode('tree')}
              className={`px-3 py-2 text-sm flex items-center gap-1 ${viewMode === 'tree' ? 'bg-white/10' : ''}`}>
              <MaterialIcon name="account_tree" className="text-sm" />
              {t('hr.orgChart.treeView')}
            </button>
            <button onClick={() => setViewMode('flat')}
              className={`px-3 py-2 text-sm flex items-center gap-1 ${viewMode === 'flat' ? 'bg-white/10' : ''}`}>
              <MaterialIcon name="list" className="text-sm" />
              {t('hr.orgChart.flatView')}
            </button>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : !filteredData ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="account_tree" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : viewMode === 'flat' ? (
        <div className="glass-card rounded-2xl p-4 sm:p-5 space-y-2">
          <OrgTreeNode node={filteredData} isFlat />
        </div>
      ) : (
        <div className="glass-card rounded-2xl p-4 sm:p-6 overflow-x-auto">
          <div className="flex justify-center min-w-[600px]">
            <OrgTreeNode node={filteredData} />
          </div>
        </div>
      )}
    </div>
  );
}
