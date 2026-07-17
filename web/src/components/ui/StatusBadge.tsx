interface StatusBadgeProps {
  label: string
  ok: boolean
  detail?: string
}

export function StatusBadge({ label, ok, detail }: StatusBadgeProps) {
  return (
    <div
      className="inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3.5 py-1.5 text-sm text-fg-muted"
      title={detail}
    >
      <span className={`h-2 w-2 rounded-full ${ok ? 'bg-bull' : 'bg-bear'}`} aria-hidden="true" />
      <span>{label}</span>
    </div>
  )
}
