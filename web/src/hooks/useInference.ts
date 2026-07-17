// Placeholder for the DL engine's rolling prediction feed.
//
// Not implemented: services/dl_engine is currently empty (no LSTM/Transformer
// forecaster, no ONNX inference, nothing writing to the `predictions` table).
// This hook is the intended landing spot for that work — see the
// Intelligence page, which renders its "not connected" state off this.
export interface InferenceState {
  connected: false
  reason: 'not_implemented'
}

export function useInference(_symbol: string): InferenceState {
  return { connected: false, reason: 'not_implemented' }
}
