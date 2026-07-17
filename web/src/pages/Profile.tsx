import { useEffect, useState } from 'react'
import { Card } from '../components/ui/Card'
import { useAuth } from '../context/AuthContext'
import {
  ApiError,
  deleteBrokerCredential,
  getBrokerCredentials,
  importEnvCredentials,
  saveBrokerCredential,
} from '../lib/api'
import type { BrokerCredentialStatus, BrokerName } from '../lib/types'
import { BROKER_LABELS, BROKER_SUBTITLES } from '../utils/format'

const BROKER_FIELDS: Record<BrokerName, { key: string; label: string; type?: string }[]> = {
  alpaca: [
    { key: 'api_key_id', label: 'API Key ID' },
    { key: 'secret_key', label: 'Secret Key', type: 'password' },
  ],
  oanda: [
    { key: 'account_id', label: 'Account ID' },
    { key: 'access_token', label: 'Access Token', type: 'password' },
  ],
  questrade: [{ key: 'refresh_token', label: 'Refresh Token', type: 'password' }],
}

export function Profile() {
  const { user } = useAuth()
  const [statuses, setStatuses] = useState<BrokerCredentialStatus[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [importMessage, setImportMessage] = useState<string | null>(null)

  async function refresh() {
    try {
      setStatuses(await getBrokerCredentials())
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load broker connections')
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  async function handleImport() {
    setImportMessage(null)
    try {
      const { imported } = await importEnvCredentials()
      setImportMessage(
        imported.length > 0
          ? `Imported: ${imported.join(', ')}`
          : 'Nothing to import — no .env broker credentials configured, or already connected',
      )
      await refresh()
    } catch (err) {
      setImportMessage(err instanceof ApiError ? err.message : 'Import failed')
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <h1 className="text-2xl font-bold text-fg">Profile</h1>

      <Card>
        <div className="text-sm text-fg-muted">Signed in as</div>
        <div className="text-lg font-semibold text-fg">{user?.email}</div>
      </Card>

      <section aria-label="Broker connections">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-fg-muted">
            Broker connections
          </h2>
          <button
            type="button"
            onClick={handleImport}
            className="rounded-lg border border-border px-3 py-1.5 text-xs text-fg-muted hover:text-fg"
          >
            Import from server .env (dev only)
          </button>
        </div>
        {importMessage && <p className="mb-3 text-sm text-fg-muted">{importMessage}</p>}
        {error && <p className="mb-3 text-sm text-bear">{error}</p>}

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {statuses?.map((status) => (
            <BrokerConnectionCard key={status.broker} status={status} onChange={refresh} />
          ))}
        </div>
      </section>
    </div>
  )
}

function BrokerConnectionCard({
  status,
  onChange,
}: {
  status: BrokerCredentialStatus
  onChange: () => void
}) {
  const [values, setValues] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleConnect() {
    setSubmitting(true)
    setError(null)
    try {
      await saveBrokerCredential(status.broker, values)
      onChange()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to save')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDisconnect() {
    setSubmitting(true)
    try {
      await deleteBrokerCredential(status.broker)
      onChange()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-base font-semibold text-fg">{BROKER_LABELS[status.broker]}</h3>
          <span className="text-sm text-fg-muted">{BROKER_SUBTITLES[status.broker]}</span>
        </div>
        <span
          className={`whitespace-nowrap rounded-full px-2.5 py-1 text-xs ${
            status.connected ? 'bg-bull/15 text-bull' : 'bg-bear/15 text-bear'
          }`}
        >
          {status.connected ? 'Connected' : 'Not connected'}
        </span>
      </div>

      {status.connected ? (
        <button
          type="button"
          onClick={handleDisconnect}
          disabled={submitting}
          className="rounded-lg border border-border px-3 py-1.5 text-sm text-fg-muted hover:text-fg disabled:opacity-50"
        >
          Disconnect
        </button>
      ) : (
        <div className="flex flex-col gap-2">
          {BROKER_FIELDS[status.broker].map((field) => (
            <input
              key={field.key}
              type={field.type ?? 'text'}
              placeholder={field.label}
              value={values[field.key] ?? ''}
              onChange={(e) => setValues((v) => ({ ...v, [field.key]: e.target.value }))}
              className="rounded-lg border border-border bg-elevated px-3 py-2 text-sm text-fg"
            />
          ))}
          {error && <p className="text-xs text-bear">{error}</p>}
          <button
            type="button"
            onClick={handleConnect}
            disabled={submitting}
            className="rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
          >
            {submitting ? 'Connecting…' : 'Connect'}
          </button>
        </div>
      )}
    </Card>
  )
}
