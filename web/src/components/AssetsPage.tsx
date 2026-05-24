import React, { useEffect, useState } from 'react';
import { api, type Asset } from '../api';
import { DataTable } from '@philiaspace/phi-dashboard';
import type { ColumnDef } from '@philiaspace/phi-dashboard';
import { Alert } from '@philiaspace/ui-primitives';
import { AssetPreview } from './AssetPreview';

export default function AssetsPage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [type, setType] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [detail, setDetail] = useState<Asset | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  useEffect(() => {
    setPage(1);
    fetchAssets(1, pageSize);
  }, [type]);

  const fetchAssets = async (p: number, ps: number) => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listAssets({ type, limit: ps, offset: (p - 1) * ps });
      setAssets(res.data || []);
      setTotal(res.total || (res.data || []).length);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch assets');
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (p: number, ps: number) => {
    setPage(p);
    setPageSize(ps);
    fetchAssets(p, ps);
  };

  const openDetail = async (row: Asset) => {
    if (row.SourceURL) {
      setDetail(row);
      return;
    }
    setDetailLoading(true);
    setDetail(null);
    try {
      const res = await api.getAsset(row.ID);
      setDetail(res.data);
    } catch {
      setDetail(row);
    } finally {
      setDetailLoading(false);
    }
  };

  const columns: ColumnDef<Asset>[] = [
    { key: 'ID', title: 'ID', render: (row) => <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{row.ID.slice(0, 16)}...</span> },
    { key: 'Type', title: 'Type' },
    { key: 'preview', title: 'Preview', render: (row) => <AssetPreview asset={row} /> },
    { key: 'QuestionID', title: 'Question', render: (row) => row.QuestionID ? <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{row.QuestionID.slice(0, 16)}...</span> : '-' },
    { key: 'PassageID', title: 'Passage', render: (row) => row.PassageID ? <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{row.PassageID.slice(0, 16)}...</span> : '-' },
    { key: 'S3Key', title: 'S3 Key', render: (row) => <span style={{ display: 'block', maxWidth: '360px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{row.S3Key}</span> },
  ];

  return (
    <div>
      <h2 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1.5rem' }}>Assets</h2>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        {[['', 'All'], ['audio', 'Audio'], ['image', 'Image']].map(([value, label]) => (
          <button key={value} onClick={() => setType(value)} style={{ padding: '0.375rem 0.75rem', background: type === value ? '#3b82f6' : '#1e293b', color: '#fff', border: '1px solid #334155', borderRadius: '6px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 600 }}>{label}</button>
        ))}
      </div>
      {error && <Alert variant="error" style={{ marginBottom: '1rem' }}>{error}</Alert>}
      <DataTable data={assets} columns={columns} loading={loading} onRowClick={openDetail} total={total} pageSize={pageSize} onPageChange={handlePageChange} />

      {detail && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem' }} onClick={() => setDetail(null)}>
          <div style={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', maxWidth: '640px', width: '100%', maxHeight: '90vh', overflow: 'auto' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ padding: '1rem 1.5rem', borderBottom: '1px solid #334155', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>Asset Detail</h3>
                <span style={{ background: detail.Type === 'audio' ? '#f59e0b' : '#06b6d4', color: '#fff', fontSize: '0.65rem', padding: '0.125rem 0.5rem', borderRadius: '9999px', fontWeight: 600 }}>{detail.Type}</span>
              </div>
              <button onClick={() => setDetail(null)} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: '1.25rem', cursor: 'pointer' }}>&times;</button>
            </div>
            <div style={{ padding: '1.5rem' }}>
              {detailLoading && <div style={{ color: '#94a3b8', textAlign: 'center', padding: '2rem' }}>Loading...</div>}
              {!detailLoading && (
                <>
                  <div style={{ marginBottom: '1.5rem' }}>
                    <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.5rem' }}>Preview</label>
                    <div style={{ background: '#0f172a', border: '1px solid #334155', borderRadius: '6px', padding: '1rem', display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '120px' }}>
                      {detail.Type === 'audio' ? (
                        <audio controls src={detail.SourceURL || `/assets/${detail.ID}`} style={{ width: '100%', maxWidth: '400px' }} />
                      ) : detail.Type === 'image' ? (
                        <img src={detail.SourceURL || `/assets/${detail.ID}`} alt={detail.ID} style={{ maxWidth: '100%', maxHeight: '400px', borderRadius: '6px' }} />
                      ) : (
                        <a href={detail.SourceURL || `/assets/${detail.ID}`} target="_blank" rel="noreferrer" style={{ color: '#60a5fa' }}>Open File</a>
                      )}
                    </div>
                  </div>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>ID</label>
                      <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8', wordBreak: 'break-all' }}>{detail.ID}</span>
                    </div>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>S3 Key</label>
                      <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8', wordBreak: 'break-all' }}>{detail.S3Key}</span>
                    </div>
                    {detail.QuestionID && (
                      <div>
                        <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>Question ID</label>
                        <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8', wordBreak: 'break-all' }}>{detail.QuestionID}</span>
                      </div>
                    )}
                    {detail.PassageID && (
                      <div>
                        <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', marginBottom: '0.25rem' }}>Passage ID</label>
                        <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8', wordBreak: 'break-all' }}>{detail.PassageID}</span>
                      </div>
                    )}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
