import type {
  BalanceResponse,
  BrokerCredentialStatus,
  CreateOrderRequest,
  HealthStatus,
  Order,
  Quote,
  User,
} from './types'

export const API_BASE_URL: string =
  import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export const logoUrl = `${API_BASE_URL}/logo.png`

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// credentials: 'include' is required on every call — the gateway's session
// cookie is HttpOnly and cross-origin (different port than the dev
// server), so without this the browser never sends or stores it.
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, { ...init, credentials: 'include' })

  if (res.status === 204) {
    return undefined as T
  }

  const body = await res.json().catch(() => null)
  if (!res.ok) {
    const message =
      body && typeof body === 'object' && 'error' in body
        ? String(body.error)
        : `${path} responded with ${res.status}`
    throw new ApiError(message, res.status)
  }
  return body as T
}

function postJSON<T>(path: string, payload: unknown): Promise<T> {
  return request<T>(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
}

export function getHealth(): Promise<HealthStatus> {
  return request<HealthStatus>('/healthz')
}

export function getBalance(): Promise<BalanceResponse> {
  return request<BalanceResponse>('/api/v1/balance')
}

export function getQuote(broker: string, symbol: string): Promise<Quote> {
  const params = new URLSearchParams({ broker, symbol })
  return request<Quote>(`/api/v1/quote?${params.toString()}`)
}

export function register(email: string, password: string): Promise<User> {
  return postJSON<User>('/api/v1/auth/register', { email, password })
}

export function login(email: string, password: string): Promise<User> {
  return postJSON<User>('/api/v1/auth/login', { email, password })
}

export function logout(): Promise<void> {
  return request<void>('/api/v1/auth/logout', { method: 'POST' })
}

export function getMe(): Promise<User> {
  return request<User>('/api/v1/auth/me')
}

export function getBrokerCredentials(): Promise<BrokerCredentialStatus[]> {
  return request<BrokerCredentialStatus[]>('/api/v1/broker-credentials')
}

export function saveBrokerCredential(
  broker: string,
  credentials: Record<string, string>,
): Promise<void> {
  return postJSON<void>('/api/v1/broker-credentials', { broker, credentials })
}

export function deleteBrokerCredential(broker: string): Promise<void> {
  return request<void>(`/api/v1/broker-credentials?broker=${encodeURIComponent(broker)}`, {
    method: 'DELETE',
  })
}

export function importEnvCredentials(): Promise<{ imported: string[] }> {
  return request<{ imported: string[] }>('/api/v1/broker-credentials/import-env', {
    method: 'POST',
  })
}

// Places a real order against one of the user's connected brokers. If that
// broker's credentials point at a live account, this executes a real trade.
export function createOrder(order: CreateOrderRequest): Promise<Order> {
  return postJSON<Order>('/api/v1/orders', order)
}

export function listOrders(): Promise<Order[]> {
  return request<Order[]>('/api/v1/orders')
}
