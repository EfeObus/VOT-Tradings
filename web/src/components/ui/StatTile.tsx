interface StatTileProps {
  label: string
  value: string
  tone?: 'accent' | 'bull' | 'bear'
}

const TONE_CLASSES: Record<NonNullable<StatTileProps['tone']>, string> = {
  accent: 'text-accent',
  bull: 'text-bull',
  bear: 'text-bear',
}

export function StatTile({ label, value, tone = 'accent' }: StatTileProps) {
  return (
    <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-5">
      <span className="text-xs uppercase tracking-wide text-fg-muted">{label}</span>
      <span className={`text-2xl font-bold ${TONE_CLASSES[tone]}`}>{value}</span>
    </div>
  )
}
