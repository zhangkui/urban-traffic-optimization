import axios from 'axios'
export const client = axios.create({ baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1', timeout: 10000 })
client.interceptors.request.use(config => { const token = localStorage.getItem('traffic-token'); if (token) config.headers.Authorization = `Bearer ${token}`; return config })
export type Page<T> = { items: T[]; total: number; page?: number; size?: number; pages?: number }
export const api = {
  summary: () => client.get('/dashboard/summary'),
  roads: (params?: Record<string, unknown>) => client.get<Page<Record<string, unknown>>>('/roads', { params }),
  intersections: () => client.get<Page<Record<string, unknown>>>('/intersections'),
  plans: (params?: Record<string, unknown>) => client.get('/timing-plans', { params }),
  readings: (params?: Record<string, unknown>) => client.get('/traffic/readings', { params }),
  events: (params?: Record<string, unknown>) => client.get('/events', { params }),
  simulations: () => client.get('/simulations'),
  createSimulation: (data: unknown) => client.post('/simulations', data),
  optimize: (data: unknown) => client.post('/optimization/candidates', data),
}
