import { Card } from '../components/ui/Card'
import { NotConnected } from '../components/ui/NotConnected'
import { usePortfolio } from '../context/PortfolioContext'
import { BROKER_LABELS, BROKER_SUBTITLES } from '../utils/format'

export function Settings() {
  const { balance, balanceError } = usePortfolio()

  return (
    <div className="flex flex-col gap-8">
      <section aria-label="API connectivity">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
          Broker API connectivity
        </h2>
        {balanceError && !balance && (
          <p className="text-sm text-bear">Couldn't reach the gateway at {balanceError.message}</p>
        )}
        <Card className="divide-y divide-border p-0">
          {balance?.brokers.map((status) => (
            <div key={status.broker} className="flex items-center justify-between gap-4 px-5 py-4">
              <div>
                <div className="font-semibold text-fg">{BROKER_LABELS[status.broker] ?? status.broker}</div>
                <div className="text-sm text-fg-muted">{BROKER_SUBTITLES[status.broker] ?? ''}</div>
              </div>
              <div className="text-right">
                <span
                  className={`rounded-full px-2.5 py-1 text-xs ${
                    status.account ? 'bg-bull/15 text-bull' : 'bg-bear/15 text-bear'
                  }`}
                >
                  {status.account ? 'Connected' : 'Unavailable'}
                </span>
                {status.error && <div className="mt-1 max-w-xs text-xs text-fg-muted">{status.error}</div>}
              </div>
            </div>
          ))}
          {!balance && <div className="px-5 py-4 text-sm text-fg-muted">Loading…</div>}
        </Card>
      </section>

      <section aria-label="Circuit breaker rules">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
          Circuit breaker rules
        </h2>
        <NotConnected
          title="Max slippage / automated execution limits"
          requires="an order-execution API to enforce against — these controls have nothing to gate yet"
        />
      </section>
    </div>
  )
}
