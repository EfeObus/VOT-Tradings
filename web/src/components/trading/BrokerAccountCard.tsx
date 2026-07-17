import type { BrokerStatus } from '../../lib/types'
import { BROKER_LABELS, BROKER_SUBTITLES, formatCurrency, formatRelativeTime } from '../../utils/format'

interface BrokerAccountCardProps {
  status: BrokerStatus
}

export function BrokerAccountCard({ status }: BrokerAccountCardProps) {
  const label = BROKER_LABELS[status.broker] ?? status.broker
  const subtitle = BROKER_SUBTITLES[status.broker] ?? ''
  const connected = Boolean(status.account)

  return (
    <div
      className={`flex flex-col gap-4 rounded-xl border bg-surface p-5 ${connected ? 'border-border' : 'border-bear/40'}`}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-base font-semibold text-fg">{label}</h3>
          <span className="mt-0.5 block text-sm text-fg-muted">{subtitle}</span>
        </div>
        <span
          className={`whitespace-nowrap rounded-full px-2.5 py-1 text-xs ${
            connected ? 'bg-bull/15 text-bull' : 'bg-bear/15 text-bear'
          }`}
        >
          {connected ? 'Connected' : 'Unavailable'}
        </span>
      </div>

      {status.account ? (
        <dl className="grid gap-2.5">
          <div className="flex justify-between gap-3 text-sm">
            <dt className="text-fg-muted">Equity</dt>
            <dd className="font-semibold text-fg">
              {formatCurrency(status.account.equity, status.account.currency)}
            </dd>
          </div>
          <div className="flex justify-between gap-3 text-sm">
            <dt className="text-fg-muted">Buying power</dt>
            <dd className="font-semibold text-fg">
              {formatCurrency(status.account.buying_power, status.account.currency)}
            </dd>
          </div>
          <div className="flex justify-between gap-3 text-sm">
            <dt className="text-fg-muted">Cash</dt>
            <dd className="font-semibold text-fg">
              {formatCurrency(status.account.cash, status.account.currency)}
            </dd>
          </div>
          {status.account.pattern_day_trader && (
            <div className="text-xs text-amber-400">Pattern Day Trader flagged</div>
          )}
          <div className="text-xs text-fg-muted">Updated {formatRelativeTime(status.account.updated_at)}</div>
        </dl>
      ) : (
        <p className="break-words text-sm text-fg-muted">{status.error}</p>
      )}
    </div>
  )
}
