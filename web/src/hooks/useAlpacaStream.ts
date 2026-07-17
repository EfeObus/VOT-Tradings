// Placeholder for real-time US equities tick ingestion.
//
// Not implemented: the gateway has no WebSocket streaming endpoint yet
// (cmd/data_pipeline is unbuilt — see root README's "Current Implementation
// Status"). This hook exists as the intended landing spot for that work so
// Market-screen components have one stable import to switch over to real
// data, rather than reaching for a fetch call themselves.
export interface StreamState {
  connected: false
  reason: 'not_implemented'
}

export function useAlpacaStream(_symbol: string): StreamState {
  return { connected: false, reason: 'not_implemented' }
}
