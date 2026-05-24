import React, { useState, useEffect } from 'react';
import { Link, useLocation, Routes, Route, Navigate } from 'react-router-dom';
import { api, type Module } from '../api';
import QuestionsPage from './QuestionsPage';
import PassagesPage from './PassagesPage';
import AssetsPage from './AssetsPage';
import TemplatesPage from './TemplatesPage';
import { Button, Card, Alert } from '@philiaspace/ui-primitives';

function LoginPage({ onLogin }: { onLogin: () => void }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await api.login(username, password);
      if (res.data?.access_token) {
        localStorage.setItem('dashboard_token', res.data.access_token);
        onLogin();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a' }}>
      <Card style={{ width: '100%', maxWidth: '360px' }}>
        <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
          <div style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>📝</div>
          <h1 style={{ fontSize: '1.25rem', fontWeight: 700, color: '#f8fafc' }}>MondaiPhi Admin</h1>
          <p style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.25rem' }}>Question Bank Dashboard</p>
        </div>
        {error && <Alert variant="error" style={{ marginBottom: '1rem' }}>{error}</Alert>}
        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.25rem' }}>Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
              style={{ width: '100%', padding: '0.5rem 0.75rem', background: '#0f172a', border: '1px solid #334155', borderRadius: '6px', color: '#f8fafc', fontSize: '0.875rem' }}
            />
          </div>
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.25rem' }}>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              style={{ width: '100%', padding: '0.5rem 0.75rem', background: '#0f172a', border: '1px solid #334155', borderRadius: '6px', color: '#f8fafc', fontSize: '0.875rem' }}
            />
          </div>
          <Button type="submit" variant="primary" disabled={loading} style={{ width: '100%' }}>
            {loading ? 'Signing in...' : 'Sign In'}
          </Button>
        </form>
        <div style={{ marginTop: '1rem', textAlign: 'center', fontSize: '0.75rem', color: '#64748b' }}>
          Default: admin / admin
        </div>
      </Card>
    </div>
  );
}

export default function Layout() {
  const [modules, setModules] = useState<Module[]>([]);
  const [title, setTitle] = useState('MondaiPhi Admin');
  const [stats, setStats] = useState<{ total_questions: number; total_passages: string; total_assets: string; total_templates: string } | null>(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [checking, setChecking] = useState(true);
  const location = useLocation();

  useEffect(() => {
    const token = localStorage.getItem('dashboard_token');
    if (token) {
      setIsLoggedIn(true);
    }
    setChecking(false);
  }, []);

  useEffect(() => {
    if (!isLoggedIn) return;
    api.getConfig().then((res) => {
      setTitle(res.data.title);
      setModules(res.data.modules);
    }).catch(console.error);
    api.getStats().then((res) => {
      setStats(res.data);
    }).catch(console.error);
  }, [isLoggedIn]);

  if (checking) {
    return <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#94a3b8' }}>Loading...</div>;
  }

  if (!isLoggedIn) {
    return <LoginPage onLogin={() => setIsLoggedIn(true)} />;
  }

  const navStyle: React.CSSProperties = {
    width: '240px',
    background: '#1e293b',
    borderRight: '1px solid #334155',
    minHeight: '100vh',
    padding: '1.5rem',
    position: 'fixed',
    left: 0,
    top: 0,
    display: 'flex',
    flexDirection: 'column',
  };

  const contentStyle: React.CSSProperties = {
    marginLeft: '240px',
    padding: '2rem',
    minHeight: '100vh',
  };

  const linkStyle = (path: string): React.CSSProperties => ({
    display: 'block',
    padding: '0.75rem 1rem',
    color: location.pathname.startsWith(path) ? '#f8fafc' : '#94a3b8',
    background: location.pathname.startsWith(path) ? '#334155' : 'transparent',
    borderRadius: '6px',
    textDecoration: 'none',
    marginBottom: '0.5rem',
    fontSize: '0.875rem',
    fontWeight: 500,
    transition: 'all 0.15s',
  });

  return (
    <div style={{ display: 'flex' }}>
      <nav style={navStyle}>
        <div style={{ marginBottom: '2rem' }}>
          <h1 style={{ fontSize: '1.125rem', fontWeight: 700, color: '#f8fafc', letterSpacing: '0.05em' }}>
            {title}
          </h1>
          <p style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.25rem' }}>
            Question Bank Admin
          </p>
        </div>
        <div style={{ flex: 1 }}>
          {modules.map((m) => (
            <Link key={m.path} to={m.path.replace('/dashboard', '')} style={linkStyle(m.path.replace('/dashboard', ''))}>
              <span style={{ marginRight: '0.5rem' }}>{m.icon}</span>
              {m.name}
            </Link>
          ))}
        </div>
        <div style={{ paddingTop: '1rem', borderTop: '1px solid #334155' }}>
          <button
            onClick={() => { localStorage.removeItem('dashboard_token'); window.location.reload(); }}
            style={{ background: 'none', border: 'none', color: '#64748b', cursor: 'pointer', fontSize: '0.75rem', width: '100%', textAlign: 'left', padding: '0.5rem' }}
          >
            🚪 Logout
          </button>
        </div>
      </nav>
      <main style={contentStyle}>
        <Routes>
          <Route path="/questions" element={<QuestionsPage />} />
          <Route path="/passages" element={<PassagesPage />} />
          <Route path="/assets" element={<AssetsPage />} />
          <Route path="/templates" element={<TemplatesPage />} />
          <Route path="/" element={<DashboardHome modules={modules} stats={stats} />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}

function DashboardHome({ modules, stats }: { modules: Module[]; stats: { total_questions: number; total_passages: string; total_assets: string; total_templates: string } | null }) {
  const statCards = stats ? [
    { label: 'Total Questions', value: stats.total_questions, icon: '📝' },
    { label: 'Passages', value: stats.total_passages, icon: '📖' },
    { label: 'Assets', value: stats.total_assets, icon: '🎵' },
    { label: 'Templates', value: stats.total_templates, icon: '📋' },
  ] : [];

  return (
    <div>
      <h2 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '1.5rem' }}>Dashboard</h2>

      {statCards.length > 0 && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
          {statCards.map((s) => (
            <Card key={s.label}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>{s.icon}</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#f8fafc' }}>{s.value}</div>
              <div style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: '0.25rem' }}>{s.label}</div>
            </Card>
          ))}
        </div>
      )}

      <h3 style={{ fontSize: '1.125rem', fontWeight: 600, marginBottom: '1rem' }}>Modules</h3>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: '1rem' }}>
        {modules.map((m) => (
          <Link
            key={m.path}
            to={m.path.replace('/dashboard', '')}
            style={{
              display: 'block',
              padding: '1.5rem',
              background: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '8px',
              textDecoration: 'none',
              color: '#f8fafc',
              transition: 'border-color 0.15s',
            }}
            onMouseEnter={(e) => (e.currentTarget.style.borderColor = '#475569')}
            onMouseLeave={(e) => (e.currentTarget.style.borderColor = '#334155')}
          >
            <div style={{ fontSize: '2rem', marginBottom: '0.75rem' }}>{m.icon}</div>
            <div style={{ fontWeight: 600, fontSize: '1rem' }}>{m.name}</div>
          </Link>
        ))}
      </div>
    </div>
  );
}
