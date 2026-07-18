import { useState, type FormEvent, type ReactNode } from 'react'
import { Card } from '../components/ui/Card'
import { NotConnected } from '../components/ui/NotConnected'
import { usePortfolio } from '../context/PortfolioContext'
import { ApiError, createOrder } from '../lib/api'
import type { BrokerName, Order, OrderSide, OrderType } from '../lib/types'
import { BROKER_LABELS, formatCurrency } from '../utils/format'

const inputClass = 'rounded-lg border border-border bg-elevated px-3 py-2 text-sm text-fg'

export function Trade() {
  const { balance } = usePortfolio()

  const cashByCurrency = new Map<string, number>()
  for (const status of balance?.brokers ?? []) {
    if (!status.account) continue
    const prev = cashByCurrency.get(status.account.currency) ?? 0
    cashByCurrency.set(status.account.currency, prev + status.account.cash)
  }

  const connectedBrokers = (balance?.brokers ?? [])
    .filter((b) => b.account)
    .map((b) => b.broker)

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">Execution Ticket</h1>

      <section aria-label="Dual-currency guardrails">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
          Available cash by currency
        </h2>
        <Card>
          {cashByCurrency.size === 0 ? (
            <p className="text-sm text-fg-muted">No connected broker accounts yet.</p>
          ) : (
            <div className="flex flex-wrap gap-6">
              {[...cashByCurrency.entries()].map(([currency, cash]) => (
                <div key={currency}>
                  <div className="text-xs uppercase tracking-wide text-fg-muted">{currency}</div>
                  <div className="text-xl font-bold text-fg">{formatCurrency(cash, currency)}</div>
                </div>
              ))}
            </div>
          )}
          <p className="mt-4 text-xs text-fg-muted">
            Real cash balances from connected brokers. The currency-friction cushion described in the
            architecture doc (50bps CAD/USD buffer) isn't computed yet — see root README.
          </p>
        </Card>
      </section>

      <section aria-label="Order ticket">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
          Order ticket
        </h2>
        {connectedBrokers.length === 0 ? (
          <NotConnected title="Order ticket" requires="a connected broker — go to Profile to connect one" />
        ) : (
          <OrderTicket brokers={connectedBrokers} />
        )}
        <p className="mt-3 text-xs text-fg-muted">
          For Alpaca orders, the FINRA Pattern Day Trader rule is checked automatically server-side
          (internal/engine/pdt.go) before the order reaches the broker — it isn't applied to OANDA or
          Questrade, which aren't subject to it. It only sees day-trades made through this app; nothing
          detects round-trips automatically yet (see root README).
        </p>
      </section>

      <NotConnected
        title="TWAP / VWAP slicing and RL-delegated routing"
        requires="the execution optimization engine described in the architecture doc — today every order routes as a single direct market/limit order"
      />
    </div>
  )
}

function OrderTicket({ brokers }: { brokers: BrokerName[] }) {
  const [broker, setBroker] = useState<BrokerName>(brokers[0])
  const [symbol, setSymbol] = useState('')
  const [side, setSide] = useState<OrderSide>('buy')
  const [type, setType] = useState<OrderType>('market')
  const [quantity, setQuantity] = useState('')
  const [limitPrice, setLimitPrice] = useState('')
  const [confirmed, setConfirmed] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<Order | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    setResult(null)
    try {
      const order = await createOrder({
        broker,
        symbol: symbol.trim().toUpperCase(),
        side,
        type,
        quantity: Number(quantity),
        limit_price: type === 'limit' ? Number(limitPrice) : undefined,
      })
      setResult(order)
      setConfirmed(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Order failed')
    } finally {
      setSubmitting(false)
    }
  }

  const canSubmit =
    confirmed &&
    symbol.trim() !== '' &&
    Number(quantity) > 0 &&
    (type !== 'limit' || Number(limitPrice) > 0)

  return (
    <Card>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <Field label="Broker">
            <select
              value={broker}
              onChange={(e) => setBroker(e.target.value as BrokerName)}
              className={inputClass}
            >
              {brokers.map((b) => (
                <option key={b} value={b}>
                  {BROKER_LABELS[b]}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Symbol">
            <input
              value={symbol}
              onChange={(e) => setSymbol(e.target.value.toUpperCase())}
              placeholder="AAPL"
              className={inputClass}
            />
          </Field>
          <Field label="Side">
            <select value={side} onChange={(e) => setSide(e.target.value as OrderSide)} className={inputClass}>
              <option value="buy">Buy</option>
              <option value="sell">Sell</option>
            </select>
          </Field>
          <Field label="Type">
            <select value={type} onChange={(e) => setType(e.target.value as OrderType)} className={inputClass}>
              <option value="market">Market</option>
              <option value="limit">Limit</option>
            </select>
          </Field>
          <Field label="Quantity">
            <input
              type="number"
              min="0"
              step="any"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              className={inputClass}
            />
          </Field>
          {type === 'limit' && (
            <Field label="Limit price">
              <input
                type="number"
                min="0"
                step="any"
                value={limitPrice}
                onChange={(e) => setLimitPrice(e.target.value)}
                className={inputClass}
              />
            </Field>
          )}
        </div>

        <label className="flex items-start gap-2 text-sm text-fg-muted">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(e) => setConfirmed(e.target.checked)}
            className="mt-1"
          />
          I understand this submits a real order to {BROKER_LABELS[broker]} — if that connection points
          at a live account, this executes with real money.
        </label>

        {error && <p className="text-sm text-bear">{error}</p>}
        {result && (
          <p className="text-sm text-bull">
            Order {result.status}
            {result.broker_order_id ? ` — broker order id ${result.broker_order_id}` : ''}
          </p>
        )}

        <button
          type="submit"
          disabled={!canSubmit || submitting}
          className="self-start rounded-lg bg-accent px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
        >
          {submitting ? 'Submitting…' : 'Place order'}
        </button>
      </form>
    </Card>
  )
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs uppercase tracking-wide text-fg-muted">{label}</label>
      {children}
    </div>
  )
}
