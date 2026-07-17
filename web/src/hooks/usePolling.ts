import { useEffect, useRef, useState } from 'react'

interface PollState<T> {
  data: T | null
  error: Error | null
  loading: boolean
}

// Polls fetcher on a fixed interval, keeping the last good `data` on screen
// if a later poll fails (a broker outage shouldn't blank the dashboard).
export function usePolling<T>(fetcher: () => Promise<T>, intervalMs: number): PollState<T> {
  const [state, setState] = useState<PollState<T>>({ data: null, error: null, loading: true })
  const fetcherRef = useRef(fetcher)
  fetcherRef.current = fetcher

  useEffect(() => {
    let cancelled = false

    async function tick() {
      try {
        const data = await fetcherRef.current()
        if (!cancelled) setState({ data, error: null, loading: false })
      } catch (err) {
        if (!cancelled) {
          setState((prev) => ({ data: prev.data, error: err as Error, loading: false }))
        }
      }
    }

    tick()
    const id = setInterval(tick, intervalMs)
    return () => {
      cancelled = true
      clearInterval(id)
    }
  }, [intervalMs])

  return state
}
