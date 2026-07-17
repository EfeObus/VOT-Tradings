import { getBalance, getHealth } from '../lib/api'
import { formatUSD } from '../lib/format'
import { usePolling } from '../hooks/usePolling'
import { BrokerCard } from './BrokerCard'
import { StatTile } from './StatTile'
import { StatusPill } from './StatusPill'

const POLL_INTERVAL_MS = 15_000

export function Dashboard() {
  const health = usePolling(getHealth, POLL_INTERVAL_MS)
  const balance = usePolling(getBalance, POLL_INTERVAL_MS)

  return (
    <main className="dashboard">
      <section className="dashboard__status-row" aria-label="System status">
        <StatusPill label="Gateway" ok={!health.error} detail={health.error?.message} />
        <StatusPill
          label="Postgres"
          ok={health.data?.postgres === 'ok'}
          detail={health.data?.postgres}
        />
        <StatusPill label="Redis" ok={health.data?.redis === 'ok'} detail={health.data?.redis} />
      </section>

      <section className="dashboard__summary" aria-label="Unified balance">
        <StatTile
          label="Total equity (USD)"
          value={formatUSD(balance.data?.unified.total_equity_usd ?? 0)}
        />
        <StatTile
          label="Total buying power (USD)"
          value={formatUSD(balance.data?.unified.total_buying_power_usd ?? 0)}
        />
        <StatTile
          label="Total cash (USD)"
          value={formatUSD(balance.data?.unified.total_cash_usd ?? 0)}
        />
      </section>

      <section className="dashboard__brokers" aria-label="Broker accounts">
        {balance.loading && !balance.data && (
          <p className="dashboard__hint">Loading broker accounts…</p>
        )}
        {balance.error && !balance.data && (
          <p className="dashboard__hint dashboard__hint--error">
            Couldn't reach the gateway at {balance.error.message}
          </p>
        )}
        {balance.data?.brokers.map((status) => (
          <BrokerCard key={status.broker} status={status} />
        ))}
      </section>
    </main>
  )
}
