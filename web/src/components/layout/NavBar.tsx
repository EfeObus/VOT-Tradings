import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { logoUrl } from '../../lib/api'

const NAV_LINKS = [
  { to: '/profile', label: 'Profile' },
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/funds', label: 'Manage funds' },
  { to: '/market/AAPL', label: 'Market data' },
  { to: '/trade', label: 'Trade' },
  { to: '/forecasts', label: 'Forecasts (AI)' },
  { to: '/reports', label: 'Reports' },
  { to: '/settings', label: 'Settings' },
]

export function NavBar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login')
  }

  return (
    <header className="border-b border-border">
      <div className="flex items-center justify-between px-8 py-5">
        <img src={logoUrl} alt="VOT Tradings" className="h-16 w-auto" />
        {user && (
          <div className="flex items-center gap-4">
            <span className="text-sm text-fg-muted">{user.email}</span>
            <button
              type="button"
              onClick={handleLogout}
              className="rounded-lg border border-border px-3 py-1.5 text-sm text-fg-muted hover:text-fg"
            >
              Log out
            </button>
          </div>
        )}
      </div>
      <nav className="flex flex-wrap gap-1 border-t border-border px-8 py-2">
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
