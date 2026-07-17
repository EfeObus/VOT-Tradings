import { useParams } from 'react-router-dom'
import { NotConnected } from '../components/ui/NotConnected'
import { useAlpacaStream } from '../hooks/useAlpacaStream'

export function Market() {
  const { symbol = '' } = useParams<{ symbol: string }>()
  const stream = useAlpacaStream(symbol)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-baseline gap-3">
        <h1 className="text-2xl font-bold text-fg">{symbol}</h1>
        <span className={`text-xs ${stream.connected ? 'text-bull' : 'text-fg-muted'}`}>
          {stream.connected ? 'Live' : 'No live feed'}
        </span>
      </div>

      <NotConnected
        title="Streaming candlesticks"
        requires="a Go WebSocket streaming service (planned: cmd/data_pipeline) — quotes today are REST-polled snapshots, not a tick stream"
      />

      <NotConnected
        title="Level 2 depth panel"
        requires="broker Level 2 order-book feeds, which none of the current brokerage drivers request"
      />

      <NotConnected
        title="Technical indicator overlay (VWAP, EMA, Bollinger Bands)"
        requires="a real tick history to compute over — nothing to chart without the streaming feed above"
      />
    </div>
  )
}
