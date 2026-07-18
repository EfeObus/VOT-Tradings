// Mirrors the JSON contracts served by the Go gateway
// (internal/httpapi/httpapi.go). Keep in sync with that file.

export interface HealthStatus {
  status: string
  postgres: string
  redis: string
}

export type BrokerName = 'alpaca' | 'oanda' | 'questrade'

export interface Account {
  id: string
  broker: BrokerName
  currency: string
  equity: number
  buying_power: number
  cash: number
  pattern_day_trader: boolean
  updated_at: string
}

export interface UnifiedBalance {
  total_equity_usd: number
  total_buying_power_usd: number
  total_cash_usd: number
  by_account: Account[] | null
}

export interface BrokerStatus {
  broker: BrokerName
  account?: Account
  equity_usd?: number
  error?: string
}

export interface BalanceResponse {
  unified: UnifiedBalance
  brokers: BrokerStatus[]
}

export interface Quote {
  broker: BrokerName
  symbol: string
  bid: number
  ask: number
  timestamp: number
}

export interface User {
  id: string
  email: string
  created_at: string
}

export interface BrokerCredentialStatus {
  broker: BrokerName
  connected: boolean
}

export type OrderSide = 'buy' | 'sell'
export type OrderType = 'market' | 'limit'

export interface CreateOrderRequest {
  broker: BrokerName
  symbol: string
  side: OrderSide
  type: OrderType
  quantity: number
  limit_price?: number
}

export interface Order {
  id: string
  user_id: string
  broker: BrokerName
  symbol: string
  side: OrderSide
  type: OrderType
  quantity: number
  limit_price?: number
  status: string
  broker_order_id?: string
  created_at: string
  updated_at: string
}
