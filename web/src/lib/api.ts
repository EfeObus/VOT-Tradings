import type { BalanceResponse, HealthStatus } from './types'

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
  if (!res.ok) {
    throw new ApiError(`${path} responded with ${res.status}`, res.status)
  }
  return (await res.json()) as T
}

export function getHealth(): Promise<HealthStatus> {
  return request<HealthStatus>('/healthz')
}

export function getBalance(): Promise<BalanceResponse> {
  return request<BalanceResponse>('/api/v1/balance')
}
