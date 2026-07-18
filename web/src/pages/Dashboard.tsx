import { Link } from 'react-router-dom'
import { AllocationDonut } from '../components/charts/AllocationDonut'
import { BrokerAccountCard } from '../components/trading/BrokerAccountCard'
import { Card } from '../components/ui/Card'
import { StatTile } from '../components/ui/StatTile'
import { usePortfolio } from '../context/PortfolioContext'
import { formatUSD } from '../utils/format'

export function Dashboard() {
  const { balance, balanceError, loading } = usePortfolio()

  const hasBrokers = (balance?.brokers.length ?? 0) > 0

  return (
    <div className="flex flex-col gap-8">
      <section className="grid grid-cols-1 gap-4 sm:grid-cols-3" aria-label="Unified NAV">
        <StatTile label="Total NAV (USD)" value={formatUSD(balance?.unified.total_equity_usd ?? 0)} />
        <StatTile
          label="Total buying power (USD)"
          value={formatUSD(balance?.unified.total_buying_power_usd ?? 0)}
          tone="bull"
        />
        <StatTile label="Total cash (USD)" value={formatUSD(balance?.unified.total_cash_usd ?? 0)} />
      </section>

      {!loading && !hasBrokers && !balanceError && (
        <Card className="flex flex-wrap items-center justify-between gap-4 border-accent/40">
          <div>
            <h2 className="text-base font-semibold text-fg">No brokers connected yet</h2>
            <p className="mt-1 text-sm text-fg-muted">
              Connect Alpaca, OANDA, or Questrade to see your real balances and allocation here.
            </p>
          </div>
          <Link
            to="/profile"
            className="whitespace-nowrap rounded-lg bg-accent px-4 py-2 text-sm font-semibold text-white"
          >
            Connect a broker
          </Link>
        </Card>
      )}

      {hasBrokers && (
        <section aria-label="Capital allocation">
          <Card>
            <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
              Allocation by broker
            </h2>
            <AllocationDonut brokers={balance?.brokers ?? []} />
          </Card>
        </section>
      )}

      <section aria-label="Cross-border split view">
        {loading && !balance && <p className="text-sm text-fg-muted">Loading broker accounts…</p>}
        {balanceError && !balance && (
          <p className="text-sm text-bear">Couldn't reach the gateway at {balanceError.message}</p>
        )}
        {hasBrokers && (
          <>
            <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
              Broker accounts
            </h2>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {balance?.brokers.map((status) => (
                <BrokerAccountCard key={status.broker} status={status} />
              ))}
            </div>
          </>
        )}
      </section>
    </div>
  )
}
