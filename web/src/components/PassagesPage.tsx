import React, { useEffect, useState } from 'react';
import { api, type Passage, type Question, type Asset } from '../api';
import { DataTable } from '@philiaspace/phi-dashboard';
import type { ColumnDef } from '@philiaspace/phi-dashboard';
import { Alert } from '@philiaspace/ui-primitives';
import { AssetPreview } from './AssetPreview';

export default function PassagesPage() {
  const [passages, setPassages] = useState<Passage[]>([]);
  const [selectedLevel, setSelectedLevel] = useState('N5');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [detail, setDetail] = useState<{ passage: Passage; questions: Question[]; assets: Asset[] } | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  useEffect(() => {
    setPage(1);
    fetchPassages(1, pageSize);
  }, [selectedLevel]);

  const fetchPassages = async (p: number, ps: number) => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listPassages({ level: selectedLevel, limit: ps, offset: (p - 1) * ps });
      setPassages(res.data || []);
      setTotal(res.total || (res.data || []).length);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch passages');
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (p: number, ps: number) => {
    setPage(p);
    setPageSize(ps);
    fetchPassages(p, ps);
  };

  const openDetail = async (row: Passage) => {
    setDetailLoading(true);
    setDetail(null);
    try {
      const res = await api.getPassage(row.ID);
      setDetail({
        passage: res.data,
        questions: (res.data as any).questions || [],
        assets: (res.data as any).assets || [],
      });
    } catch {
      setDetail({ passage: row, questions: [], assets: [] });
    } finally {
      setDetailLoading(false);
    }
  };

  const columns: ColumnDef<Passage>[] = [
    { key: 'ID', title: 'ID', render: (row) => <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{row.ID.slice(0, 16)}...</span> },
    { key: 'Level', title: 'Level' },
    { key: 'Section', title: 'Section' },
    { key: 'PassageNumber', title: 'No.' },
    { key: 'Title', title: 'Title', render: (row) => row.Title || '-' },
    { key: 'Content', title: 'Content', render: (row) => <span style={{ display: 'block', maxWidth: '520px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{row.Content}</span> },
  ];

  return (
    <div>
      <h2 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1.5rem' }}>Passages</h2>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        {['N5', 'N4', 'N3', 'N2', 'N1'].map((level) => (
          <button key={level} onClick={() => setSelectedLevel(level)} style={{ padding: '0.375rem 0.75rem', background: selectedLevel === level ? '#3b82f6' : '#1e293b', color: '#fff', border: '1px solid #334155', borderRadius: '6px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 600 }}>{level}</button>
        ))}
      </div>
      {error && <Alert variant="error" style={{ marginBottom: '1rem' }}>{error}</Alert>}
      <DataTable data={passages} columns={columns} loading={loading} onRowClick={openDetail} total={total} pageSize={pageSize} onPageChange={handlePageChange} />

      {detail && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem' }} onClick={() => setDetail(null)}>
          <div style={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', maxWidth: '720px', width: '100%', maxHeight: '90vh', overflow: 'auto' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ padding: '1rem 1.5rem', borderBottom: '1px solid #334155', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>Passage Detail</h3>
                <span style={{ background: '#3b82f6', color: '#fff', fontSize: '0.65rem', padding: '0.125rem 0.5rem', borderRadius: '9999px', fontWeight: 600 }}>{detail.passage.Level}</span>
                <span style={{ background: '#8b5cf6', color: '#fff', fontSize: '0.65rem', padding: '0.125rem 0.5rem', borderRadius: '9999px', fontWeight: 600 }}>{detail.passage.Section}</span>
              </div>
              <button onClick={() => setDetail(null)} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: '1.25rem', cursor: 'pointer' }}>&times;</button>
            </div>
            <div style={{ padding: '1.5rem' }}>
              {detailLoading && <div style={{ color: '#94a3b8', textAlign: 'center', padding: '2rem' }}>Loading...</div>}
              {!detailLoading && (
                <>
                  <div style={{ marginBottom: '1rem' }}>
                    <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>ID</label>
                    <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{detail.passage.ID}</span>
                  </div>
                  {detail.passage.Title && (
                    <div style={{ marginBottom: '1rem' }}>
                      <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>Title</label>
                      <div style={{ color: '#f8fafc', fontWeight: 600 }}>{detail.passage.Title}</div>
                    </div>
                  )}
                  <div style={{ marginBottom: '1.5rem' }}>
                    <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>Content</label>
                    <div style={{ background: '#0f172a', border: '1px solid #334155', borderRadius: '6px', padding: '1rem', color: '#f8fafc', fontSize: '0.875rem', lineHeight: '1.6', whiteSpace: 'pre-wrap' }}>{detail.passage.Content}</div>
                  </div>
                  {detail.assets.length > 0 && (
                    <div style={{ marginBottom: '1.5rem' }}>
                      <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.5rem' }}>Assets ({detail.assets.length})</label>
                      <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
                        {detail.assets.map((a) => <AssetPreview key={a.ID} asset={a} />)}
                      </div>
                    </div>
                  )}
                  {detail.questions.length > 0 && (
                    <div>
                      <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.5rem' }}>Questions ({detail.questions.length})</label>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                        {detail.questions.map((q, i) => (
                          <div key={q.ID} style={{ background: '#0f172a', border: '1px solid #334155', borderRadius: '6px', padding: '0.75rem' }}>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', marginBottom: '0.5rem' }}>
                              <span style={{ fontSize: '0.7rem', color: '#64748b', fontWeight: 600 }}>Q{i + 1}</span>
                              <span style={{ color: '#22c55e', fontSize: '0.7rem', fontWeight: 600 }}>Ans: {q.AnswerValue}</span>
                            </div>
                            <div style={{ color: '#f8fafc', fontSize: '0.8rem', lineHeight: '1.5' }}>{q.Prompt}</div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
