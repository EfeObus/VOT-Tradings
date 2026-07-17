import { NotConnected } from '../components/ui/NotConnected'

export function Intelligence() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold text-fg">AI Inference Telemetry</h1>

      <NotConnected
        title="LSTM forecasting matrix (T+5 / T+15 / T+60)"
        requires="the Python DL engine (services/dl_engine is currently empty — no model, no ONNX runtime, no inference API)"
      />

      <NotConnected
        title="Model confidence / RMSE tracking"
        requires="predictions actually being written to the predictions table — the schema exists, nothing populates it yet"
      />

      <NotConnected
        title="Self-correction log"
        requires="the offline fine-tuning pipeline described in the architecture doc, which hasn't been built"
      />
    </div>
  )
}
