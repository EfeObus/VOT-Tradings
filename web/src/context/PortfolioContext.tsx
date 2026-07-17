import { createContext, useContext, type ReactNode } from 'react'
import { usePolling } from '../hooks/usePolling'
import { getBalance, getHealth } from '../lib/api'
import type { BalanceResponse, HealthStatus } from '../lib/types'

const POLL_INTERVAL_MS = 15_000

interface PortfolioState {
  health: HealthStatus | null
  healthError: Error | null
  balance: BalanceResponse | null
  balanceError: Error | null
  loading: boolean
}

const PortfolioContext = createContext<PortfolioState | null>(null)

// Single source of truth for live account state, polled once and shared by
// every page (Dashboard, Settings, Trade) instead of each page hitting the
// gateway on its own timer.
export function PortfolioProvider({ children }: { children: ReactNode }) {
  const health = usePolling(getHealth, POLL_INTERVAL_MS)
  const balance = usePolling(getBalance, POLL_INTERVAL_MS)

  const value: PortfolioState = {
    health: health.data,
    healthError: health.error,
    balance: balance.data,
    balanceError: balance.error,
    loading: health.loading || balance.loading,
  }

  return <PortfolioContext.Provider value={value}>{children}</PortfolioContext.Provider>
}

export function usePortfolio(): PortfolioState {
  const ctx = useContext(PortfolioContext)
  if (!ctx) {
    throw new Error('usePortfolio must be used within a PortfolioProvider')
  }
  return ctx
}
