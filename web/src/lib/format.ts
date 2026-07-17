const usdFormatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' })

export function formatUSD(value: number): string {
  return usdFormatter.format(value)
}

const currencyFormatters = new Map<string, Intl.NumberFormat>()

export function formatCurrency(value: number, currency: string): string {
  let formatter = currencyFormatters.get(currency)
  if (!formatter) {
    formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency })
    currencyFormatters.set(currency, formatter)
  }
  return formatter.format(value)
}

export function formatRelativeTime(iso: string): string {
  const diffSec = Math.round((Date.now() - new Date(iso).getTime()) / 1000)
  if (diffSec < 5) return 'just now'
  if (diffSec < 60) return `${diffSec}s ago`
  const diffMin = Math.round(diffSec / 60)
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.round(diffMin / 60)
  return `${diffHr}h ago`
}

export const BROKER_LABELS: Record<string, string> = {
  alpaca: 'Alpaca',
  oanda: 'OANDA',
  questrade: 'Questrade',
}

export const BROKER_SUBTITLES: Record<string, string> = {
  alpaca: 'US Equities',
  oanda: 'Forex',
  questrade: 'Canadian Equities',
}
