import type { BrokerStatus } from '../lib/types'
import { BROKER_LABELS, BROKER_SUBTITLES, formatCurrency, formatRelativeTime } from '../lib/format'

interface BrokerCardProps {
  status: BrokerStatus
}

export function BrokerCard({ status }: BrokerCardProps) {
  const label = BROKER_LABELS[status.broker] ?? status.broker
  const subtitle = BROKER_SUBTITLES[status.broker] ?? ''
  const connected = Boolean(status.account)

  return (
    <div className={`broker-card ${connected ? 'is-connected' : 'is-error'}`}>
      <div className="broker-card__header">
        <div>
          <h3>{label}</h3>
          <span className="broker-card__subtitle">{subtitle}</span>
        </div>
        <span className={`broker-card__badge ${connected ? 'is-connected' : 'is-error'}`}>
          {connected ? 'Connected' : 'Unavailable'}
        </span>
      </div>

      {status.account ? (
        <dl className="broker-card__stats">
          <div>
            <dt>Equity</dt>
            <dd>{formatCurrency(status.account.equity, status.account.currency)}</dd>
          </div>
          <div>
            <dt>Buying power</dt>
            <dd>{formatCurrency(status.account.buying_power, status.account.currency)}</dd>
          </div>
          <div>
            <dt>Cash</dt>
            <dd>{formatCurrency(status.account.cash, status.account.currency)}</dd>
          </div>
          {status.account.pattern_day_trader && (
            <div className="broker-card__flag">Pattern Day Trader flagged</div>
          )}
          <div className="broker-card__updated">Updated {formatRelativeTime(status.account.updated_at)}</div>
        </dl>
      ) : (
        <p className="broker-card__error">{status.error}</p>
      )}
    </div>
  )
}
