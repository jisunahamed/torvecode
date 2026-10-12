import { describe, expect, test } from "bun:test"
import { torveLogo } from "../src/logo"

describe("Torve terminal logo", () => {
  test("keeps the full logo free of ANSI controls before OpenTUI renders it", () => {
    const logo = torveLogo(120, 48)

    expect(logo).not.toMatch(/\u001b/)
    expect(logo.split("\n").length).toBeLessThanOrEqual(36)
    expect(logo).toContain("▆")
  })

  test("uses the compact Torve wordmark when the terminal cannot fit the artwork", () => {
    expect(torveLogo(120, 37)).toBe("TORVE AI")
    expect(torveLogo(79, 48)).toBe("TORVE AI")
  })
})
