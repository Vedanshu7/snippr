import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true, // send session cookie on every request
})

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      // Clear persisted auth state and redirect to login.
      localStorage.removeItem('snippr-auth')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export interface Snippet {
  id: number
  user_id: number
  workspace_id?: number
  title: string
  content: string
  language: string
  is_public: boolean
  share_slug?: string
  tags: string[]
  created_at: string
  updated_at: string
}

export interface User {
  id: number
  email: string
}

export interface Workspace {
  id: number
  owner_id: number
  name: string
  invite_code: string
  is_public: boolean
  created_at: string
}

export interface WorkspaceMember {
  workspace_id: number
  user_id: number
  email: string
  role: string
  joined_at: string
}

export interface CreateSnippetReq {
  title: string
  content: string
  language: string
  is_public: boolean
  tags: string[]
  workspace_id?: number
}

export const auth = {
  register: (email: string, password: string) =>
    api.post('/register', { email, password }),
  login: (email: string, password: string) =>
    api.post<{ token: string; user: User }>('/login', { email, password }),
  logout: () => api.post('/logout'),
  me: () => api.get<User>('/me'),
}

export const snippets = {
  list: (q?: string, tag?: string, workspaceId?: number) =>
    api.get<Snippet[]>('/snippets', { params: { q, tag, workspace_id: workspaceId } }),
  get: (id: number) => api.get<Snippet>(`/snippets/${id}`),
  create: (data: CreateSnippetReq) => api.post<Snippet>('/snippets', data),
  update: (id: number, data: CreateSnippetReq) =>
    api.put<Snippet>(`/snippets/${id}`, data),
  delete: (id: number) => api.delete(`/snippets/${id}`),
  getPublic: (slug: string) =>
    axios.get<Snippet>(`/api/s/${slug}`),
}

export const workspaces = {
  list: () => api.get<Workspace[]>('/workspaces'),
  listPublic: () => axios.get<Workspace[]>('/api/workspaces/public', { withCredentials: true }),
  get: (id: number) => api.get<Workspace>(`/workspaces/${id}`),
  create: (name: string, is_public: boolean) =>
    api.post<Workspace>('/workspaces', { name, is_public }),
  update: (id: number, data: { name: string; is_public: boolean }) =>
    api.put<Workspace>(`/workspaces/${id}`, data),
  join: (invite_code: string) => api.post<Workspace>('/workspaces/join', { invite_code }),
  joinPublic: (workspace_id: number) =>
    api.post<Workspace>('/workspaces/join', { workspace_id }),
  leave: (id: number) => api.delete(`/workspaces/${id}/leave`),
  delete: (id: number) => api.delete(`/workspaces/${id}`),
  members: (id: number) => api.get<WorkspaceMember[]>(`/workspaces/${id}/members`),
  removeMember: (workspaceId: number, userId: number) =>
    api.delete(`/workspaces/${workspaceId}/members/${userId}`),
  rotateInvite: (id: number) =>
    api.post<{ invite_code: string }>(`/workspaces/${id}/rotate-invite`),
}
