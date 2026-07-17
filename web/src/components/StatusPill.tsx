interface StatusPillProps {
  label: string
  ok: boolean
  detail?: string
}

export function StatusPill({ label, ok, detail }: StatusPillProps) {
  return (
    <div className={`status-pill ${ok ? 'is-ok' : 'is-down'}`} title={detail}>
      <span className="status-pill__dot" aria-hidden="true" />
      <span>{label}</span>
    </div>
  )
}
