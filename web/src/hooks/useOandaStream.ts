import { useEffect, useState } from 'react'
import { API_BASE_URL } from '../lib/api'

export interface StreamTick {
  broker: string
  symbol: string
  bid: number
  ask: number
  timestamp: number
}

export interface StreamState {
  connected: boolean
  tick: StreamTick | null
  error: string | null
}

// Real live price stream over the gateway's authenticated WebSocket
// endpoint (GET /ws/quotes), backed by OANDA's actual v20 streaming API —
// see internal/brokerage/oanda's StreamPricing. The session cookie is sent
// automatically on the WS handshake by the browser; no credentials flag is
// needed the way fetch requires one.
//
// This delivers live bid/ask ticks, not OHLC candlesticks — turning a tick
// stream into time-bucketed bars for a chart is separate, still-unbuilt
// work (see the Market page's remaining NotConnected panels).
export function useOandaStream(symbol: string): StreamState {
  const [state, setState] = useState<StreamState>({ connected: false, tick: null, error: null })

  useEffect(() => {
    if (!symbol) return

    const wsUrl = `${API_BASE_URL.replace(/^http/, 'ws')}/ws/quotes?broker=oanda&symbol=${encodeURIComponent(symbol)}`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => setState((s) => ({ ...s, connected: true, error: null }))
    ws.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as StreamTick | { error: string }
        if ('error' in parsed) {
          setState((s) => ({ ...s, error: parsed.error }))
          return
        }
        setState((s) => ({ ...s, tick: parsed, error: null }))
      } catch {
        // ignore unparseable frames
      }
    }
    ws.onerror = () => setState((s) => ({ ...s, error: 'stream connection error' }))
    ws.onclose = () => setState((s) => ({ ...s, connected: false }))

    return () => ws.close()
  }, [symbol])

  return state
}
