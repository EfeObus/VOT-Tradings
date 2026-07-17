import { useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card } from '../components/ui/Card'
import { NotConnected } from '../components/ui/NotConnected'
import { ApiError, getQuote } from '../lib/api'
import type { BrokerName, Quote } from '../lib/types'
import { BROKER_LABELS } from '../utils/format'

const BROKERS: BrokerName[] = ['alpaca', 'oanda', 'questrade']

export function Market() {
  const { symbol = 'AAPL' } = useParams<{ symbol: string }>()
  const navigate = useNavigate()

  const [broker, setBroker] = useState<BrokerName>('alpaca')
  const [symbolInput, setSymbolInput] = useState(symbol)
  const [quote, setQuote] = useState<Quote | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const trimmed = symbolInput.trim()
    if (!trimmed) return

    navigate(`/market/${trimmed}`, { replace: true })
    setLoading(true)
    setError(null)
    try {
      setQuote(await getQuote(broker, trimmed))
    } catch (err) {
      setQuote(null)
      setError(err instanceof ApiError ? err.message : 'Failed to fetch quote')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">Market data</h1>

      <Card>
        <form onSubmit={handleSubmit} className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1">
            <label htmlFor="broker" className="text-xs uppercase tracking-wide text-fg-muted">
              Broker
            </label>
            <select
              id="broker"
              value={broker}
              onChange={(e) => setBroker(e.target.value as BrokerName)}
              className="rounded-lg border border-border bg-elevated px-3 py-2 text-sm text-fg"
            >
              {BROKERS.map((b) => (
                <option key={b} value={b}>
                  {BROKER_LABELS[b]}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label htmlFor="symbol" className="text-xs uppercase tracking-wide text-fg-muted">
              Symbol
            </label>
            <input
              id="symbol"
              value={symbolInput}
              onChange={(e) => setSymbolInput(e.target.value.toUpperCase())}
              placeholder="AAPL, EUR_USD, ..."
              className="rounded-lg border border-border bg-elevated px-3 py-2 text-sm text-fg"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="rounded-lg bg-accent px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
          >
            {loading ? 'Fetching…' : 'Get quote'}
          </button>
        </form>

        {error && <p className="mt-4 text-sm text-bear">{error}</p>}

        {quote && (
          <div className="mt-5 flex flex-wrap gap-8 border-t border-border pt-5">
            <div>
              <div className="text-xs uppercase tracking-wide text-fg-muted">Bid</div>
              <div className="text-xl font-bold text-bear">{quote.bid}</div>
            </div>
            <div>
              <div className="text-xs uppercase tracking-wide text-fg-muted">Ask</div>
              <div className="text-xl font-bold text-bull">{quote.ask}</div>
            </div>
            <div>
              <div className="text-xs uppercase tracking-wide text-fg-muted">As of</div>
              <div className="text-xl font-bold text-fg">
                {new Date(quote.timestamp).toLocaleTimeString()}
              </div>
            </div>
          </div>
        )}

        <p className="mt-4 text-xs text-fg-muted">
          Real REST snapshot from the broker — click "Get quote" again to refresh. Not a live stream.
        </p>
      </Card>

      <NotConnected
        title="Streaming candlesticks"
        requires="a Go WebSocket streaming service (planned: cmd/data_pipeline) — the lookup above is a REST snapshot, not a tick stream"
      />

      <NotConnected
        title="Level 2 depth panel"
        requires="broker Level 2 order-book feeds, which none of the current brokerage drivers request"
      />

      <NotConnected
        title="Technical indicator overlay (VWAP, EMA, Bollinger Bands)"
        requires="a real tick history to compute over — nothing to chart without the streaming feed above"
      />
    </div>
  )
}
