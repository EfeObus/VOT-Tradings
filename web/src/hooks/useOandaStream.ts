// Placeholder for real-time Forex tick ingestion. See useAlpacaStream.ts —
// same situation, no backend streaming endpoint exists yet.
export interface StreamState {
  connected: false
  reason: 'not_implemented'
}

export function useOandaStream(_symbol: string): StreamState {
  return { connected: false, reason: 'not_implemented' }
}
