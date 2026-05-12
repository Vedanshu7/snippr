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

export interface CreateSnippetReq {
  title: string
  content: string
  language: string
  is_public: boolean
  tags: string[]
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
  list: (q?: string, tag?: string) =>
    api.get<Snippet[]>('/snippets', { params: { q, tag } }),
  get: (id: number) => api.get<Snippet>(`/snippets/${id}`),
  create: (data: CreateSnippetReq) => api.post<Snippet>('/snippets', data),
  update: (id: number, data: CreateSnippetReq) =>
    api.put<Snippet>(`/snippets/${id}`, data),
  delete: (id: number) => api.delete(`/snippets/${id}`),
  getPublic: (slug: string) =>
    axios.get<Snippet>(`/api/s/${slug}`),
}
