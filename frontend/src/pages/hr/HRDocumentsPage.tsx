import { useState, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const TYPE_ICONS: Record<string, { icon: string; color: string }> = {
  contract: { icon: 'description', color: 'text-blue-400' },
  policy: { icon: 'policy', color: 'text-green-400' },
  certificate: { icon: 'workspace_premium', color: 'text-amber-400' },
  id: { icon: 'badge', color: 'text-purple-400' },
  tax: { icon: 'receipt_long', color: 'text-red-400' },
  other: { icon: 'folder', color: 'text-gray-400' },
};

export function HRDocumentsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [filters, setFilters] = useState({ type: '', search: '' });
  const [showUpload, setShowUpload] = useState(false);
  const [uploadForm, setUploadForm] = useState({ title: '', type: 'contract', description: '' });
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['hr-documents', filters],
    queryFn: () => api.hr.getDocuments(filters),
  });

  const documents: Record<string, unknown>[] = data?.data || data || [];

  const uploadMutation = useMutation({
    mutationFn: (formData: FormData) => api.hr.uploadDocument(formData),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-documents'] }); setShowUpload(false); setSelectedFile(null); },
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.hr.deleteDocument(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-documents'] }),
  });

  const handleUpload = () => {
    if (!selectedFile) return;
    const formData = new FormData();
    formData.append('file', selectedFile);
    formData.append('title', uploadForm.title || selectedFile.name);
    formData.append('type', uploadForm.type);
    formData.append('description', uploadForm.description);
    uploadMutation.mutate(formData);
  };

  const handleDownload = async (id: string, fileName: string) => {
    try {
      const blob = await api.hr.downloadDocument(id);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      // エラー無視
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const filteredDocs = documents.filter(d => {
    if (filters.type && d.type !== filters.type) return false;
    if (filters.search && !(d.title as string || '').toLowerCase().includes(filters.search.toLowerCase())) return false;
    return true;
  });

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.documents.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.documents.subtitle')}</p>
          </div>
        </div>
        <button onClick={() => setShowUpload(true)}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
          <MaterialIcon name="upload" className="text-lg" />
          {t('hr.documents.upload')}
        </button>
      </div>

      {/* 検索 & フィルター */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <MaterialIcon name="search" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground text-xl" />
          <input value={filters.search} onChange={(e) => setFilters(f => ({...f, search: e.target.value}))}
            placeholder={t('common.search')}
            className="w-full pl-10 pr-4 py-2.5 glass-input rounded-xl text-sm" />
        </div>
        <select value={filters.type} onChange={(e) => setFilters(f => ({...f, type: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[140px]">
          <option value="">{t('hr.documents.documentType')}</option>
          {['contract', 'policy', 'certificate', 'id', 'tax', 'other'].map(tp => (
            <option key={tp} value={tp}>{t(`hr.documents.types.${tp}`)}</option>
          ))}
        </select>
      </div>

      {/* ドキュメントリスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : filteredDocs.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="description" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <>
          {/* モバイルカード */}
          <div className="md:hidden space-y-3">
            {filteredDocs.map((doc) => {
              const typeInfo = TYPE_ICONS[(doc.type as string) || 'other'] || TYPE_ICONS.other;
              return (
                <div key={doc.id as string} className="glass-card rounded-2xl p-4">
                  <div className="flex items-start gap-3">
                    <div className={`size-10 rounded-xl flex items-center justify-center shrink-0 ${typeInfo.color} bg-white/5`}>
                      <MaterialIcon name={typeInfo.icon} className="text-xl" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-semibold text-sm truncate">{String(doc.title)}</p>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                        <span>{t(`hr.documents.types.${doc.type || 'other'}`)}</span>
                        {!!doc.file_size && <span>· {formatFileSize(doc.file_size as number)}</span>}
                      </div>
                      {!!doc.description && <p className="text-xs text-muted-foreground mt-1 line-clamp-1">{String(doc.description)}</p>}
                      <p className="text-[10px] text-muted-foreground mt-1">{String(doc.uploaded_at || doc.created_at || '')}</p>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <button onClick={() => handleDownload(doc.id as string, doc.file_name as string || doc.title as string)}
                        className="p-2 hover:bg-white/10 rounded-lg transition-colors">
                        <MaterialIcon name="download" className="text-base text-indigo-400" />
                      </button>
                      <button onClick={() => { if (confirm(t('common.confirm'))) deleteMutation.mutate(doc.id as string); }}
                        className="p-2 hover:bg-red-500/10 rounded-lg transition-colors">
                        <MaterialIcon name="delete" className="text-base text-red-400" />
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* デスクトップテーブル */}
          <div className="hidden md:block glass-card rounded-2xl overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/10 text-xs text-muted-foreground">
                  <th className="text-left px-5 py-3 font-medium">{t('hr.documents.fileName')}</th>
                  <th className="text-left px-5 py-3 font-medium">{t('hr.documents.documentType')}</th>
                  <th className="text-left px-5 py-3 font-medium">{t('hr.documents.fileSize')}</th>
                  <th className="text-left px-5 py-3 font-medium">{t('hr.documents.uploadDate')}</th>
                  <th className="text-right px-5 py-3 font-medium">{t('common.actions')}</th>
                </tr>
              </thead>
              <tbody>
                {filteredDocs.map((doc) => {
                  const typeInfo = TYPE_ICONS[(doc.type as string) || 'other'] || TYPE_ICONS.other;
                  return (
                    <tr key={doc.id as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="px-5 py-3">
                        <div className="flex items-center gap-3">
                          <MaterialIcon name={typeInfo.icon} className={`text-lg ${typeInfo.color}`} />
                          <div>
                            <p className="text-sm font-medium">{String(doc.title)}</p>
                            {!!doc.description && <p className="text-xs text-muted-foreground truncate max-w-[300px]">{String(doc.description)}</p>}
                          </div>
                        </div>
                      </td>
                      <td className="px-5 py-3 text-sm">{t(`hr.documents.types.${doc.type || 'other'}`)}</td>
                      <td className="px-5 py-3 text-sm text-muted-foreground">{doc.file_size ? formatFileSize(doc.file_size as number) : '-'}</td>
                      <td className="px-5 py-3 text-sm text-muted-foreground">{String(doc.uploaded_at || doc.created_at || '-')}</td>
                      <td className="px-5 py-3">
                        <div className="flex items-center justify-end gap-1">
                          <button onClick={() => handleDownload(doc.id as string, doc.file_name as string || doc.title as string)}
                            className="p-1.5 hover:bg-white/10 rounded-lg transition-colors" title={t('hr.documents.download')}>
                            <MaterialIcon name="download" className="text-base text-indigo-400" />
                          </button>
                          <button onClick={() => { if (confirm(t('common.confirm'))) deleteMutation.mutate(doc.id as string); }}
                            className="p-1.5 hover:bg-red-500/10 rounded-lg transition-colors">
                            <MaterialIcon name="delete" className="text-base text-red-400" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* アップロードモーダル */}
      {showUpload && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.documents.upload')}</h2>
              <button onClick={() => { setShowUpload(false); setSelectedFile(null); }}
                className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              {/* ドロップゾーン */}
              <div onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-white/20 rounded-xl p-6 text-center cursor-pointer hover:border-indigo-400/50 hover:bg-white/5 transition-all">
                <input ref={fileInputRef} type="file" className="hidden"
                  onChange={(e) => { if (e.target.files?.[0]) setSelectedFile(e.target.files[0]); }} />
                {selectedFile ? (
                  <div className="space-y-1">
                    <MaterialIcon name="check_circle" className="text-3xl text-green-400" />
                    <p className="text-sm font-medium">{selectedFile.name}</p>
                    <p className="text-xs text-muted-foreground">{formatFileSize(selectedFile.size)}</p>
                  </div>
                ) : (
                  <div className="space-y-1">
                    <MaterialIcon name="cloud_upload" className="text-3xl text-muted-foreground" />
                    <p className="text-sm text-muted-foreground">{t('hr.documents.dropzone')}</p>
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.documents.fileName')}</label>
                <input value={uploadForm.title} onChange={(e) => setUploadForm(f => ({...f, title: e.target.value}))}
                  placeholder={selectedFile?.name || ''}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.documents.documentType')}</label>
                <select value={uploadForm.type} onChange={(e) => setUploadForm(f => ({...f, type: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                  {['contract', 'policy', 'certificate', 'id', 'tax', 'other'].map(tp => (
                    <option key={tp} value={tp}>{t(`hr.documents.types.${tp}`)}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.documents.description')}</label>
                <textarea value={uploadForm.description} onChange={(e) => setUploadForm(f => ({...f, description: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={handleUpload}
                disabled={!selectedFile || uploadMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('hr.documents.upload')}
              </button>
              <button onClick={() => { setShowUpload(false); setSelectedFile(null); }}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
