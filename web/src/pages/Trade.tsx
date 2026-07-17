import { Card } from '../components/ui/Card'
import { NotConnected } from '../components/ui/NotConnected'
import { usePortfolio } from '../context/PortfolioContext'
import { formatCurrency } from '../utils/format'

export function Trade() {
  const { balance } = usePortfolio()

  const cashByCurrency = new Map<string, number>()
  for (const status of balance?.brokers ?? []) {
    if (!status.account) continue
    const prev = cashByCurrency.get(status.account.currency) ?? 0
    cashByCurrency.set(status.account.currency, prev + status.account.cash)
  }

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

      <NotConnected
        title="Order ticket (Market / Limit / TWAP / VWAP / RL-delegated routing)"
        requires="an order-execution HTTP endpoint — PlaceOrder exists per-broker in internal/brokerage/ but the gateway doesn't expose it over HTTP yet"
      />

      <NotConnected
        title="Pattern Day Trader risk shield"
        requires="exposing internal/engine/pdt.go's CheckPDT over the HTTP API — the rule is implemented server-side, nothing calls it yet"
      />
    </div>
  )
}
