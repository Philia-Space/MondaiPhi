const API_BASE = '/dashboard/api';

export interface ApiResponse<T> {
  data: T;
  count?: number;
  total?: number;
  message?: string;
}

export interface Question {
  ID: string;
  Level: string;
  Section: string;
  Prompt: string;
  Context: string;
  AnswerValue: string;
  AnswerNote: string;
  PassageID: string;
  SourceGroupKey: string;
  version: number;
}

export interface Option {
  ID: string;
  QuestionID: string;
  Value: string;
  Label: string;
  SortOrder: number;
}

export interface Passage {
  ID: string;
  PassageNumber: number;
  Title: string;
  Content: string;
  Level: string;
  Section: string;
}

export interface Asset {
  ID: string;
  Type: string;
  SourceURL: string;
  S3Key: string;
  LocalPath: string;
  QuestionID: string;
  PassageID: string;
}

export interface Template {
  ID: string;
  Name: string;
  Level: string;
  SectionCounts: Record<string, number>;
  TotalQuestions: number;
  IsDefault: boolean;
}

export interface Module {
  name: string;
  path: string;
  icon: string;
  priority: number;
}

export interface DashboardConfig {
  title: string;
  prefix: string;
  require_auth: boolean;
  modules: Module[];
}

function getToken(): string | null {
  return localStorage.getItem('dashboard_token');
}

async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(url, {
    ...options,
    headers: {
      ...headers,
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `HTTP ${res.status}` }));
    throw new Error(err.message || err.error?.message || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  login: (username: string, password: string) =>
    fetchJson<ApiResponse<{ access_token: string; token_type: string; expires_in: number }>>(`${API_BASE}/login`, {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  getConfig: () => fetchJson<ApiResponse<DashboardConfig>>(`${API_BASE}/config`),
  getModules: () => fetchJson<ApiResponse<Module[]>>(`${API_BASE}/modules`),
  getStats: () => fetchJson<ApiResponse<{ total_questions: number; total_passages: string; total_assets: string; total_templates: string }>>(`${API_BASE}/stats`),

  listQuestions: (params?: { level?: string; section?: string; search?: string; sort?: string; sort_dir?: string; limit?: number; offset?: number }) => {
    const qs = new URLSearchParams();
    if (params?.level) qs.set('level', params.level);
    if (params?.section) qs.set('section', params.section);
    if (params?.search) qs.set('search', params.search);
    if (params?.sort) qs.set('sort', params.sort);
    if (params?.sort_dir) qs.set('sort_dir', params.sort_dir);
    if (params?.limit) qs.set('limit', String(params.limit));
    if (params?.offset) qs.set('offset', String(params.offset));
    return fetchJson<ApiResponse<Question[]>>(`${API_BASE}/questions?${qs}`);
  },
  getQuestion: (id: string) => fetchJson<ApiResponse<{ question: Question; options: Option[] }>>(`${API_BASE}/questions/${id}`),
  getQuestionAssets: (id: string) => fetchJson<ApiResponse<Asset[]>>(`${API_BASE}/questions/${id}/assets`),
  batchQuestionAssets: (ids: string[]) =>
    fetchJson<ApiResponse<Record<string, Asset[]>>>(`${API_BASE}/questions/assets/batch`, {
      method: 'POST',
      body: JSON.stringify({ question_ids: ids }),
    }),
  createQuestion: (body: Record<string, unknown>) =>
    fetchJson<ApiResponse<{ id: string; message: string }>>(`${API_BASE}/questions`, { method: 'POST', body: JSON.stringify(body) }),
  updateQuestion: (id: string, body: Record<string, unknown>) =>
    fetchJson<ApiResponse<{ id: string; message: string }>>(`${API_BASE}/questions/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteQuestion: (id: string) =>
    fetchJson<ApiResponse<{ id: string; message: string }>>(`${API_BASE}/questions/${id}`, { method: 'DELETE' }),

  listPassages: (params?: { level?: string; section?: string; limit?: number; offset?: number }) => {
    const qs = new URLSearchParams();
    if (params?.level) qs.set('level', params.level);
    if (params?.section) qs.set('section', params.section);
    if (params?.limit) qs.set('limit', String(params.limit));
    if (params?.offset) qs.set('offset', String(params.offset));
    return fetchJson<ApiResponse<Passage[]>>(`${API_BASE}/passages?${qs}`);
  },
  getPassage: (id: string) =>
    fetchJson<ApiResponse<Passage & { questions: Question[]; assets: Asset[] }>>(`${API_BASE}/passages/${id}`),

  listAssets: (params?: { type?: string; limit?: number; offset?: number }) => {
    const qs = new URLSearchParams();
    if (params?.type) qs.set('type', params.type);
    if (params?.limit) qs.set('limit', String(params.limit));
    if (params?.offset) qs.set('offset', String(params.offset));
    return fetchJson<ApiResponse<Asset[]>>(`${API_BASE}/assets?${qs}`);
  },
  getAsset: (id: string) => fetchJson<ApiResponse<Asset>>(`${API_BASE}/assets/${id}`),

  listTemplates: (level?: string) => {
    const qs = level ? `?level=${level}` : '';
    return fetchJson<ApiResponse<Template[]>>(`${API_BASE}/templates${qs}`);
  },
};
