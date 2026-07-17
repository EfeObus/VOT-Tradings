import { NavLink } from 'react-router-dom'
import { logoUrl } from '../../lib/api'

const NAV_LINKS = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/market/AAPL', label: 'Market' },
  { to: '/intelligence', label: 'Intelligence' },
  { to: '/trade', label: 'Trade' },
  { to: '/settings', label: 'Settings' },
]

export function NavBar() {
  return (
    <header className="flex flex-wrap items-center gap-6 border-b border-border px-8 py-4">
      <img src={logoUrl} alt="VOT Tradings" className="h-9 w-auto" />
      <nav className="flex flex-wrap gap-1">
        {NAV_LINKS.map((link) => (
          <NavLink
            key={link.to}
            to={link.to}
            className={({ isActive }) =>
              `rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
                isActive ? 'bg-elevated text-fg' : 'text-fg-muted hover:text-fg'
              }`
            }
          >
            {link.label}
          </NavLink>
        ))}
      </nav>
    </header>
  )
}
