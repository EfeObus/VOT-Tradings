import type { BalanceResponse, HealthStatus, Quote } from './types'

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

async function request<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`)
  const body = await res.json().catch(() => null)
  if (!res.ok) {
    const message = body && typeof body === 'object' && 'error' in body ? String(body.error) : `${path} responded with ${res.status}`
    throw new ApiError(message, res.status)
  }
  return body as T
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
