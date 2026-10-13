import { describe, expect, test } from "bun:test"
import { isTorveCodingModel, torveModelKey } from "../../src/plugin/torve-models"

const model = (id: string, tools = true) => ({
  id,
  status: "active",
  capabilities: { tools, input: ["text"], output: ["text"] },
})

describe("Torve model catalog", () => {
  test("keeps coding models with tool support", () => {
    expect(isTorveCodingModel(model("gpt-5.6-sol"))).toBe(true)
    expect(isTorveCodingModel(model("gpt-5.6-sol", false))).toBe(false)
  })

  test("removes non-coding gateway entries even when their metadata claims tool support", () => {
    expect(isTorveCodingModel(model("google-diffusiongemma-26b-a4b-it"))).toBe(false)
    expect(isTorveCodingModel(model("nvidia-llama-3.1-nemoguard-8b-content-safety"))).toBe(false)
  })

  test("normalizes aliases before duplicate detection", () => {
    expect(torveModelKey({ modelID: " GPT-5.6-Sol " })).toBe("gpt-5.6-sol")
    expect(torveModelKey({ modelID: "gpt-5.6-sol" })).toBe("gpt-5.6-sol")
  })
})
