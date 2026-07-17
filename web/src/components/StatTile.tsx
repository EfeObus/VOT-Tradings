interface StatTileProps {
  label: string
  value: string
}

export function StatTile({ label, value }: StatTileProps) {
  return (
    <div className="stat-tile">
      <span className="stat-tile__label">{label}</span>
      <span className="stat-tile__value">{value}</span>
    </div>
  )
}
