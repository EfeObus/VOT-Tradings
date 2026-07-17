import { logoUrl } from '../lib/api'

export function Header() {
  return (
    <header className="header">
      <img src={logoUrl} alt="VOT Tradings" className="header__logo" />
    </header>
  )
}
