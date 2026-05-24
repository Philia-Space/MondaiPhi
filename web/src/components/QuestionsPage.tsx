import React, { useState, useEffect } from 'react';
import { api, type Asset, type Option, type Question } from '../api';
import { DataTable, CrudForm } from '@philiaspace/phi-dashboard';
import type { ColumnDef, FormField } from '@philiaspace/phi-dashboard';
import { Button, Card, Alert } from '@philiaspace/ui-primitives';
import { AssetPreview } from './AssetPreview';

const labelStyle: React.CSSProperties = { fontSize: '0.75rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.25rem' };
const valueStyle: React.CSSProperties = { fontSize: '0.875rem', color: '#f8fafc', marginBottom: '1rem' };

function QuestionDetail({ question, options, assets, onClose, onEdit, onDelete }: {
  question: Question;
  options: Option[];
  assets: Asset[];
  onClose: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div style={{
      position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem',
    }}>
      <div style={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', maxWidth: '640px', width: '100%', maxHeight: '90vh', overflow: 'auto' }}>
        <div style={{ padding: '1rem', borderBottom: '1px solid #334155', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>Question Preview</h3>
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <Button variant="ghost" size="sm" onClick={onEdit}>Edit</Button>
            <Button variant="danger" size="sm" onClick={onDelete}>Delete</Button>
            <button onClick={onClose} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: '1.25rem', cursor: 'pointer', marginLeft: '0.5rem' }}>&times;</button>
          </div>
        </div>
        <div style={{ padding: '1.25rem' }}>
          <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
            <span style={{ padding: '0.25rem 0.5rem', background: '#3b82f620', color: '#60a5fa', borderRadius: '4px', fontSize: '0.75rem', fontWeight: 600 }}>{question.Level}</span>
            <span style={{ padding: '0.25rem 0.5rem', background: '#8b5cf620', color: '#a78bfa', borderRadius: '4px', fontSize: '0.75rem', fontWeight: 600 }}>{question.Section}</span>
          </div>

          <div style={labelStyle}>ID</div>
          <div style={{ ...valueStyle, fontFamily: 'monospace', fontSize: '0.75rem', color: '#64748b' }}>{question.ID}</div>

          {question.PassageID && (
            <>
              <div style={labelStyle}>Passage</div>
              <div style={{ ...valueStyle, fontFamily: 'monospace', fontSize: '0.75rem', color: '#64748b' }}>{question.PassageID}</div>
            </>
          )}

          <div style={labelStyle}>Prompt</div>
          <div style={{ ...valueStyle, whiteSpace: 'pre-wrap', lineHeight: 1.6 }}>{question.Prompt}</div>

          {question.Context && (
            <>
              <div style={labelStyle}>Context</div>
              <div style={{ ...valueStyle, whiteSpace: 'pre-wrap', lineHeight: 1.6, color: '#cbd5e1' }}>{question.Context}</div>
            </>
          )}

          {options.length > 0 && (
            <>
              <div style={labelStyle}>Options</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1rem' }}>
                {options.sort((a, b) => a.SortOrder - b.SortOrder).map((opt) => (
                  <div key={opt.ID || opt.Value} style={{
                    display: 'flex', alignItems: 'center', gap: '0.75rem',
                    padding: '0.5rem 0.75rem', background: opt.Value === question.AnswerValue ? '#22c55e10' : '#0f172a',
                    border: `1px solid ${opt.Value === question.AnswerValue ? '#22c55e40' : '#1e293b'}`,
                    borderRadius: '6px',
                  }}>
                    <span style={{
                      width: '24px', height: '24px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: '0.75rem', fontWeight: 700, flexShrink: 0,
                      background: opt.Value === question.AnswerValue ? '#22c55e' : '#334155',
                      color: opt.Value === question.AnswerValue ? '#fff' : '#94a3b8',
                    }}>{opt.Value}</span>
                    <span style={{ fontSize: '0.875rem', color: '#f8fafc', flex: 1 }}>{opt.Label}</span>
                    {opt.Value === question.AnswerValue && (
                      <span style={{ fontSize: '0.75rem', color: '#22c55e', fontWeight: 600 }}>Correct</span>
                    )}
                  </div>
                ))}
              </div>
            </>
          )}

          {question.AnswerNote && (
            <>
              <div style={labelStyle}>Answer Note</div>
              <div style={{ ...valueStyle, color: '#cbd5e1', whiteSpace: 'pre-wrap' }}>{question.AnswerNote}</div>
            </>
          )}

          {assets.length > 0 && (
            <>
              <div style={labelStyle}>Assets</div>
              <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '1rem' }}>
                {assets.map((a) => (
                  <div key={a.ID} style={{ padding: '0.5rem', background: '#0f172a', borderRadius: '6px', border: '1px solid #334155' }}>
                    <div style={{ fontSize: '0.625rem', color: '#64748b', marginBottom: '0.25rem' }}>{a.Type}</div>
                    <AssetPreview asset={a} />
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default function QuestionsPage() {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedLevel, setSelectedLevel] = useState('N5');
  const [selectedSection, setSelectedSection] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortCol, setSortCol] = useState('');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | 'preview' | null>(null);
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null);
  const [previewOptions, setPreviewOptions] = useState<Option[]>([]);
  const [previewAssets, setPreviewAssets] = useState<Asset[]>([]);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [assetsByQuestion, setAssetsByQuestion] = useState<Record<string, Asset[]>>({});
  const [refreshKey, setRefreshKey] = useState(0);

  const fetchData = async (p: number, ps: number, lvl: string, sec: string, q: string, sc: string, sd: string) => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listQuestions({ level: lvl, section: sec, search: q, sort: sc, sort_dir: sd, limit: ps, offset: (p - 1) * ps });
      const items = res.data || [];
      setQuestions(items);
      setTotal(res.total || items.length);
      if (items.length > 0) {
        const batchRes = await api.batchQuestionAssets(items.map((qq) => qq.ID)).catch(() => ({ data: {} as Record<string, Asset[]> }));
        setAssetsByQuestion(batchRes.data || {});
      } else {
        setAssetsByQuestion({});
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData(page, pageSize, selectedLevel, selectedSection, searchQuery, sortCol, sortDir);
  }, [selectedLevel, selectedSection, searchQuery, sortCol, sortDir, page, pageSize, refreshKey]);

  const handlePageChange = (p: number, ps: number) => {
    setPage(p);
    setPageSize(ps);
  };

  const handleSearch = (q: string) => {
    setSearchQuery(q);
    if (page !== 1) setPage(1);
  };

  const handleSort = (col: string, dir: 'asc' | 'desc') => {
    setSortCol(col);
    setSortDir(dir);
    if (page !== 1) setPage(1);
  };

  const openPreview = async (row: Question) => {
    setEditingQuestion(row);
    setPreviewOptions([]);
    setPreviewAssets(assetsByQuestion[row.ID] || []);
    setModalMode('preview');
    try {
      const res = await api.getQuestion(row.ID);
      setPreviewOptions(res.data?.options || []);
    } catch {
      // options not critical for preview
    }
  };

  const handleDelete = async (row: Question) => {
    if (!confirm(`Delete question ${row.ID.slice(0, 16)}...?`)) return;
    try {
      await api.deleteQuestion(row.ID);
      setQuestions((prev) => prev.filter((q) => q.ID !== row.ID));
      setModalMode(null);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  const handleEdit = (row: Question) => {
    setEditingQuestion(row);
    setModalMode('edit');
    setFormError(null);
  };

  const handleCreate = () => {
    setEditingQuestion(null);
    setModalMode('create');
    setFormError(null);
  };

  const handleFormSubmit = async (data: Record<string, unknown>) => {
    setFormLoading(true);
    setFormError(null);
    try {
      if (modalMode === 'edit' && editingQuestion) {
        await api.updateQuestion(editingQuestion.ID, {
          prompt: data.prompt,
          context: data.context,
          answer_value: data.answer_value,
          answer_note: data.answer_note,
        });
      } else if (modalMode === 'create') {
        await api.createQuestion({
          level: data.level,
          section: data.section,
          prompt: data.prompt,
          context: data.context,
          answer_value: data.answer_value,
          answer_note: data.answer_note,
          options: [
            { value: '1', label: data.option_1 as string || 'Option 1', sort_order: 0 },
            { value: '2', label: data.option_2 as string || 'Option 2', sort_order: 1 },
            { value: '3', label: data.option_3 as string || 'Option 3', sort_order: 2 },
            { value: '4', label: data.option_4 as string || 'Option 4', sort_order: 3 },
          ],
        });
      }
      setModalMode(null);
      setEditingQuestion(null);
      setRefreshKey((k) => k + 1);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setFormLoading(false);
    }
  };

  const columns: ColumnDef<Question>[] = [
    { key: 'ID', title: 'ID', sortable: true, sortKey: 'id', render: (row) => <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#94a3b8' }}>{row.ID.slice(0, 16)}...</span> },
    { key: 'Level', title: 'Level', sortable: true, sortKey: 'level' },
    { key: 'Section', title: 'Section', sortable: true, sortKey: 'section' },
    { key: 'Prompt', title: 'Prompt', render: (row) => <span style={{ display: 'block', maxWidth: '400px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{row.Prompt}</span> },
    { key: 'answer_value', title: 'Answer', sortable: true, sortKey: 'answer_value', render: (row) => <span style={{ color: '#22c55e', fontWeight: 600 }}>{row.AnswerValue}</span> },
    { key: 'assets', title: 'Preview', render: (row) => {
      const a = assetsByQuestion[row.ID] || [];
      if (!a.length) return <span style={{ color: '#64748b' }}>-</span>;
      return <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
        {a.slice(0, 2).map((asset) => <AssetPreview key={asset.ID} asset={asset} />)}
        {a.length > 2 && <span style={{ color: '#94a3b8', fontSize: '0.75rem' }}>+{a.length - 2}</span>}
      </div>;
    }},
  ];

  const editFields: FormField[] = [
    { key: 'prompt', label: 'Prompt', type: 'textarea', required: true, rows: 3 },
    { key: 'context', label: 'Context', type: 'textarea', rows: 2 },
    { key: 'answer_value', label: 'Answer Value', type: 'text', required: true },
    { key: 'answer_note', label: 'Answer Note', type: 'textarea', rows: 2 },
  ];

  const createFields: FormField[] = [
    { key: 'level', label: 'Level', type: 'select', required: true, options: [
      { value: 'N5', label: 'N5' }, { value: 'N4', label: 'N4' }, { value: 'N3', label: 'N3' },
      { value: 'N2', label: 'N2' }, { value: 'N1', label: 'N1' },
    ]},
    { key: 'section', label: 'Section', type: 'select', required: true, options: [
      { value: 'grammar', label: 'Grammar' }, { value: 'reading', label: 'Reading' }, { value: 'listening', label: 'Listening' },
    ]},
    { key: 'prompt', label: 'Prompt', type: 'textarea', required: true, rows: 3 },
    { key: 'context', label: 'Context', type: 'textarea', rows: 2 },
    { key: 'answer_value', label: 'Answer Value', type: 'text', required: true },
    { key: 'answer_note', label: 'Answer Note', type: 'textarea', rows: 2 },
    { key: 'option_1', label: 'Option 1', type: 'text', required: true },
    { key: 'option_2', label: 'Option 2', type: 'text', required: true },
    { key: 'option_3', label: 'Option 3', type: 'text', required: true },
    { key: 'option_4', label: 'Option 4', type: 'text', required: true },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Questions</h2>
        <Button variant="primary" onClick={handleCreate}>+ New Question</Button>
      </div>

      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
        {['N5', 'N4', 'N3', 'N2', 'N1'].map((level) => (
          <button
            key={level}
            onClick={() => setSelectedLevel(level)}
            style={{
              padding: '0.375rem 0.75rem',
              background: selectedLevel === level ? '#3b82f6' : '#1e293b',
              color: '#fff',
              border: '1px solid #334155',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.75rem',
              fontWeight: 600,
            }}
          >
            {level}
          </button>
        ))}
        <span style={{ color: '#475569', margin: '0 0.25rem' }}>|</span>
        {[['all', '', 'All'], ['grammar', 'grammar', 'Grammar'], ['reading', 'reading', 'Reading'], ['listening', 'listening', 'Listening']].map(([key, value, label]) => (
          <button
            key={key}
            onClick={() => setSelectedSection(value)}
            style={{
              padding: '0.375rem 0.75rem',
              background: selectedSection === value ? '#8b5cf6' : '#1e293b',
              color: '#fff',
              border: '1px solid #334155',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.75rem',
              fontWeight: 600,
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {error && <Alert variant="error" style={{ marginBottom: '1rem' }}>{error}</Alert>}

      <DataTable
        data={questions}
        columns={columns}
        loading={loading}
        onRowClick={openPreview}
        onEdit={handleEdit}
        onDelete={handleDelete}
        total={total}
        pageSize={pageSize}
        onPageChange={handlePageChange}
        onSearch={handleSearch}
        onSort={handleSort}
        sortColumn={sortCol}
        sortDirection={sortDir}
      />

      {modalMode === 'preview' && editingQuestion && (
        <QuestionDetail
          question={editingQuestion}
          options={previewOptions}
          assets={previewAssets}
          onClose={() => { setModalMode(null); setEditingQuestion(null); }}
          onEdit={() => { setModalMode('edit'); setFormError(null); }}
          onDelete={() => handleDelete(editingQuestion)}
        />
      )}

      {(modalMode === 'edit' || modalMode === 'create') && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem',
        }}>
          <div style={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', maxWidth: '500px', width: '100%', maxHeight: '90vh', overflow: 'auto' }}>
            <div style={{ padding: '1rem', borderBottom: '1px solid #334155', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 700 }}>{modalMode === 'edit' ? 'Edit Question' : 'Create Question'}</h3>
              <button onClick={() => { setModalMode(null); setEditingQuestion(null); }} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: '1.25rem', cursor: 'pointer' }}>&times;</button>
            </div>
            <div style={{ padding: '1rem' }}>
              {formError && <Alert variant="error" style={{ marginBottom: '1rem' }}>{formError}</Alert>}
              <CrudForm
                fields={modalMode === 'edit' ? editFields : createFields}
                initialData={modalMode === 'edit' && editingQuestion ? {
                  prompt: editingQuestion.Prompt,
                  context: editingQuestion.Context,
                  answer_value: editingQuestion.AnswerValue,
                  answer_note: editingQuestion.AnswerNote,
                } : undefined}
                onSubmit={handleFormSubmit}
                onCancel={() => { setModalMode(null); setEditingQuestion(null); }}
                loading={formLoading}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
