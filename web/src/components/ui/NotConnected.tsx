interface NotConnectedProps {
  title: string
  requires: string
}

// Renders an explicit "this isn't real yet" state. Used anywhere a page
// describes a feature (streaming charts, AI inference, order execution)
// that the gateway doesn't back yet, so the UI never shows fabricated
// numbers dressed up as live data.
export function NotConnected({ title, requires }: NotConnectedProps) {
  return (
    <div className="flex flex-col items-start gap-2 rounded-xl border border-dashed border-border bg-surface/50 p-6">
      <span className="text-sm font-semibold text-fg">{title}</span>
      <span className="text-sm text-fg-muted">Not connected — requires {requires}.</span>
    </div>
  )
}
