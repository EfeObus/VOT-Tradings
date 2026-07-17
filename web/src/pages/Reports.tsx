import { NotConnected } from '../components/ui/NotConnected'

export function Reports() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">Reports</h1>

      <NotConnected
        title="Order / trade history"
        requires="orders actually being persisted — the orders table exists in the schema, but nothing writes to it since order placement isn't exposed over HTTP yet (see Trade page)"
      />
      <NotConnected
        title="Performance & P/L statements"
        requires="trade history above, plus historical price data to mark positions against"
      />
    </div>
  )
}
