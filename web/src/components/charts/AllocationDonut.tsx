import { BROKER_LABELS, formatUSD } from '../../utils/format'
import type { BrokerStatus } from '../../lib/types'

interface AllocationDonutProps {
  brokers: BrokerStatus[]
}

const SLICE_COLORS = ['var(--color-accent)', 'var(--color-bull)', '#a78bfa']

const SIZE = 160
const STROKE = 22
const RADIUS = (SIZE - STROKE) / 2
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

// Real per-broker USD equity split, computed from the gateway's equity_usd
// field (internal/httpapi) — never a client-side guess at conversion rates.
export function AllocationDonut({ brokers }: AllocationDonutProps) {
  const slices = brokers.filter(
    (b): b is BrokerStatus & { equity_usd: number } =>
      typeof b.equity_usd === 'number' && b.equity_usd > 0,
  )
  const total = slices.reduce((sum, b) => sum + b.equity_usd, 0)

  let offset = 0

  return (
    <div className="flex items-center gap-6">
      <svg width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} role="img" aria-label="Capital allocation by broker">
        <circle
          cx={SIZE / 2}
          cy={SIZE / 2}
          r={RADIUS}
          fill="none"
          stroke="var(--color-elevated)"
          strokeWidth={STROKE}
        />
        {total > 0 &&
          slices.map((slice, i) => {
            const fraction = slice.equity_usd / total
            const dash = fraction * CIRCUMFERENCE
            const circle = (
              <circle
                key={slice.broker}
                cx={SIZE / 2}
                cy={SIZE / 2}
                r={RADIUS}
                fill="none"
                stroke={SLICE_COLORS[i % SLICE_COLORS.length]}
                strokeWidth={STROKE}
                strokeDasharray={`${dash} ${CIRCUMFERENCE - dash}`}
                strokeDashoffset={-offset}
                transform={`rotate(-90 ${SIZE / 2} ${SIZE / 2})`}
              />
            )
            offset += dash
            return circle
          })}
      </svg>

      <div className="flex flex-col gap-2">
        {total === 0 && <span className="text-sm text-fg-muted">No equity to allocate yet.</span>}
        {slices.map((slice, i) => (
          <div key={slice.broker} className="flex items-center gap-2 text-sm">
            <span
              className="h-2.5 w-2.5 rounded-full"
              style={{ background: SLICE_COLORS[i % SLICE_COLORS.length] }}
              aria-hidden="true"
            />
            <span className="text-fg">{BROKER_LABELS[slice.broker] ?? slice.broker}</span>
            <span className="text-fg-muted">{formatUSD(slice.equity_usd)}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
