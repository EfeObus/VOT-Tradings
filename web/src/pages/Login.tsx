import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Card } from '../components/ui/Card'
import { useAuth } from '../context/AuthContext'
import { ApiError, logoUrl } from '../lib/api'

export function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await login(email, password)
      navigate('/dashboard')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Login failed')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-canvas px-4">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <img src={logoUrl} alt="VOT Tradings" className="mx-auto h-16 w-auto" />
        <Card>
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <h1 className="text-lg font-semibold text-fg">Log in</h1>
            <div className="flex flex-col gap-1">
              <label htmlFor="email" className="text-xs uppercase tracking-wide text-fg-muted">
                Email
              </label>
              <input
                id="email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="rounded-lg border border-border bg-elevated px-3 py-2 text-sm text-fg"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label htmlFor="password" className="text-xs uppercase tracking-wide text-fg-muted">
                Password
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="rounded-lg border border-border bg-elevated px-3 py-2 text-sm text-fg"
              />
            </div>
            {error && <p className="text-sm text-bear">{error}</p>}
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-accent px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
            >
              {submitting ? 'Logging in…' : 'Log in'}
            </button>
            <p className="text-center text-sm text-fg-muted">
              No account? <Link to="/register" className="text-accent">Register</Link>
            </p>
          </form>
        </Card>
      </div>
    </div>
  )
}
