import { define } from "@opencode/plugin/effect/plugin"
import { Effect } from "effect"

export const TORVE_PROVIDER_ID = "torveai"

const unsupported = new Set([
  "gemma-2-2b",
  "google-diffusiongemma-26b-a4b-it",
  "google-gemma-4-31b-it-2",
  "meta-llama-3.2-11b-vision-instruct-2",
  "nvidia-ising-calibration-1.5-31b-2",
  "nvidia-llama-3.1-nemoguard-8b-content-safety",
  "nvidia-llama-3.1-nemotron-safety-guard-8b-v3",
  "nvidia-nemotron-3.5-content-safety-2",
])

export function torveModelKey(model: { modelID: string }) {
  return model.modelID.trim().toLowerCase()
}

export function isTorveCodingModel(model: {
  id: string
  status: string
  capabilities: { tools: boolean; input: readonly string[]; output: readonly string[] }
}) {
  if (unsupported.has(model.id.toLowerCase())) return false
  if (model.status === "deprecated") return false
  if (!model.capabilities.tools) return false
  if (!model.capabilities.input.some((item) => item.startsWith("text"))) return false
  return model.capabilities.output.some((item) => item.startsWith("text"))
}

export const TorveModelsPlugin = define({
  id: "torve.models",
  effect: Effect.fn(function* (ctx) {
    yield* ctx.provider.transform((providers) => {
      const torve = providers.get(TORVE_PROVIDER_ID)
      if (!torve) return

      for (const definition of providers.list()) {
        if (definition.provider.id !== TORVE_PROVIDER_ID) providers.remove(definition.provider.id)
      }
      for (const model of torve.models.values()) {
        if (!isTorveCodingModel(model)) providers.models.remove(TORVE_PROVIDER_ID, model.id)
      }
      providers.update(TORVE_PROVIDER_ID, (provider) => {
        provider.name = "Torve AI"
      })
    })

    yield* ctx.model.transform((models) => {
      if (!models.provider.get(TORVE_PROVIDER_ID)) return

      const seen = new Set<string>()
      for (const model of models.list()) {
        if (model.providerID !== TORVE_PROVIDER_ID || !isTorveCodingModel(model)) {
          models.remove(model.providerID, model.id)
          continue
        }
        const canonical = torveModelKey(model)
        if (seen.has(canonical)) {
          models.remove(model.providerID, model.id)
          continue
        }
        seen.add(canonical)
      }

      const current = models.default.get()
      if (current?.providerID === TORVE_PROVIDER_ID && models.get(current.providerID, current.modelID)) return
      const available = models.list(TORVE_PROVIDER_ID)
      const preferred = ["gpt-5.6-sol", "claude-sonnet-5", "deepseek-v4-pro", "gemini-3-1-pro"]
        .flatMap((id) => available.find((model) => model.id === id) ?? [])
        .at(0)
      const fallback = preferred ?? available.at(0)
      if (fallback) models.default.set(TORVE_PROVIDER_ID, fallback.id)
    })
  }),
})
