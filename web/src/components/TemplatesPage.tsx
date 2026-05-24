import React, { useState, useEffect } from 'react';
import { api, type Template } from '../api';
import { DataTable } from '@philiaspace/phi-dashboard';
import type { ColumnDef } from '@philiaspace/phi-dashboard';
import { Alert } from '@philiaspace/ui-primitives';

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchTemplates();
  }, []);

  const fetchTemplates = async () => {
    setLoading(true);
    try {
      const res = await api.listTemplates();
      setTemplates(res.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnDef<Template>[] = [
    { key: 'ID', title: 'ID', render: (row) => <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{row.ID.slice(0, 16)}...</span> },
    { key: 'Name', title: 'Name' },
    { key: 'Level', title: 'Level' },
    { key: 'TotalQuestions', title: 'Questions' },
    { key: 'IsDefault', title: 'Default', render: (row) => row.IsDefault ? 'Yes' : 'No' },
  ];

  return (
    <div>
      <h2 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1.5rem' }}>Templates</h2>
      {error && <Alert variant="error" style={{ marginBottom: '1rem' }}>{error}</Alert>}
      <DataTable data={templates} columns={columns} loading={loading} />
    </div>
  );
}
