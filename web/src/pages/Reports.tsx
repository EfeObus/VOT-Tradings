import { useEffect, useState } from 'react'
import { Card } from '../components/ui/Card'
import { NotConnected } from '../components/ui/NotConnected'
import { ApiError, listOrders } from '../lib/api'
import type { Order } from '../lib/types'
import { BROKER_LABELS } from '../utils/format'

export function Reports() {
  const [orders, setOrders] = useState<Order[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    listOrders()
      .then(setOrders)
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Failed to load orders'))
  }, [])

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">Reports</h1>

      <section aria-label="Order history">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-fg-muted">
          Order history
        </h2>
        {error && <p className="text-sm text-bear">{error}</p>}
        {orders && orders.length === 0 && (
          <p className="text-sm text-fg-muted">
            No orders placed through this app yet — see the Trade page.
          </p>
        )}
        {orders && orders.length > 0 && (
          <Card className="overflow-x-auto p-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-fg-muted">
                  <th className="px-4 py-3">Time</th>
                  <th className="px-4 py-3">Broker</th>
                  <th className="px-4 py-3">Symbol</th>
                  <th className="px-4 py-3">Side</th>
                  <th className="px-4 py-3">Type</th>
                  <th className="px-4 py-3 text-right">Quantity</th>
                  <th className="px-4 py-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr key={order.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3 text-fg-muted">
                      {new Date(order.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-fg">{BROKER_LABELS[order.broker]}</td>
                    <td className="px-4 py-3 text-fg">{order.symbol}</td>
                    <td className={`px-4 py-3 ${order.side === 'buy' ? 'text-bull' : 'text-bear'}`}>
                      {order.side}
                    </td>
                    <td className="px-4 py-3 text-fg-muted">{order.type}</td>
                    <td className="px-4 py-3 text-right text-fg">{order.quantity}</td>
                    <td className="px-4 py-3 text-fg-muted">{order.status}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}
      </section>

      <NotConnected
        title="Performance & P/L statements"
        requires="historical price data to mark positions against — order history above is real, P/L attribution isn't built yet"
      />
    </div>
  )
}
