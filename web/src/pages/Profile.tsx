import { NotConnected } from '../components/ui/NotConnected'

export function Profile() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">Profile</h1>

      <div className="rounded-xl border border-bear/40 bg-bear/10 p-5 text-sm text-fg">
        <strong className="block text-bear">No authentication exists yet.</strong>
        <p className="mt-2 text-fg-muted">
          The gateway has zero user accounts or session handling — every endpoint is open to
          anyone who can reach it, gated only by network access. There's no register/login screen
          here because a cosmetic one with nothing behind it would be worse than none: it would
          imply security that doesn't exist. Building this for real means a users table, password
          hashing, sessions, and — since the app currently assumes one shared set of broker
          credentials in <code>.env</code> — a redesign to per-user encrypted broker credentials
          before login means anything.
        </p>
      </div>

      <NotConnected
        title="Register / login"
        requires="a real auth backend (users table, password hashing, sessions) — not implemented"
      />
      <NotConnected
        title="Per-user broker connections"
        requires="per-user encrypted API key/OAuth storage — today the gateway uses one shared set of credentials for everyone"
      />
    </div>
  )
}
