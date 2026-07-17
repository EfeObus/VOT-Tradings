import { Card } from '../components/ui/Card'

const FUNDING_PORTALS: { broker: string; subtitle: string; url: string }[] = [
  { broker: 'Alpaca', subtitle: 'US Equities', url: 'https://app.alpaca.markets' },
  { broker: 'OANDA', subtitle: 'Forex', url: 'https://hub.oanda.com' },
  { broker: 'Questrade', subtitle: 'Canadian Equities', url: 'https://login.questrade.com' },
]

export function Funds() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">Manage funds</h1>

      <div className="rounded-xl border border-border bg-surface/50 p-5 text-sm text-fg-muted">
        Deposits and withdrawals aren't reimplemented here — they go through each broker's own
        portal below. That's not a gap unique to this app: Alpaca, OANDA, and Questrade all keep
        fund movement off their public trading APIs for KYC/AML reasons, so every third-party
        trading app links out the same way.
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {FUNDING_PORTALS.map((p) => (
          <Card key={p.broker} className="flex flex-col gap-3">
            <div>
              <h3 className="text-base font-semibold text-fg">{p.broker}</h3>
              <span className="text-sm text-fg-muted">{p.subtitle}</span>
            </div>
            <a
              href={p.url}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-auto rounded-lg bg-accent px-4 py-2 text-center text-sm font-semibold text-white"
            >
              Open {p.broker} →
            </a>
          </Card>
        ))}
      </div>
    </div>
  )
}
