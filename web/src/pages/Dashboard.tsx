import { AllocationDonut } from '../components/charts/AllocationDonut'
import { BrokerAccountCard } from '../components/trading/BrokerAccountCard'
import { Card } from '../components/ui/Card'
import { StatTile } from '../components/ui/StatTile'
import { StatusBadge } from '../components/ui/StatusBadge'
import { usePortfolio } from '../context/PortfolioContext'
import { formatUSD } from '../utils/format'

export function Dashboard() {
  const { health, healthError, balance, balanceError, loading } = usePortfolio()

  return (
    <div className="flex flex-col gap-8">
      <section className="flex flex-wrap gap-3" aria-label="System status">
        <StatusBadge label="Gateway" ok={!healthError} detail={healthError?.message} />
        <StatusBadge label="Postgres" ok={health?.postgres === 'ok'} detail={health?.postgres} />
        <StatusBadge label="Redis" ok={health?.redis === 'ok'} detail={health?.redis} />
      </section>

      <section className="grid grid-cols-1 gap-4 sm:grid-cols-3" aria-label="Unified NAV">
        <StatTile label="Total NAV (USD)" value={formatUSD(balance?.unified.total_equity_usd ?? 0)} />
        <StatTile
          label="Total buying power (USD)"
          value={formatUSD(balance?.unified.total_buying_power_usd ?? 0)}
          tone="bull"
        />
        <StatTile label="Total cash (USD)" value={formatUSD(balance?.unified.total_cash_usd ?? 0)} />
      </section>

      <section aria-label="Capital allocation">
        <Card>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
            Allocation by broker
          </h2>
          <AllocationDonut brokers={balance?.brokers ?? []} />
        </Card>
      </section>

      <section aria-label="Cross-border split view">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
          Broker accounts
        </h2>
        {loading && !balance && <p className="text-sm text-fg-muted">Loading broker accounts…</p>}
        {balanceError && !balance && (
          <p className="text-sm text-bear">Couldn't reach the gateway at {balanceError.message}</p>
        )}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {balance?.brokers.map((status) => (
            <BrokerAccountCard key={status.broker} status={status} />
          ))}
        </div>
      </section>
    </div>
  )
}
