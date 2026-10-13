import { describe, expect, test } from "bun:test"
import { torveLogo } from "../src/logo"

describe("Torve terminal logo", () => {
  test("keeps the full logo free of ANSI controls before OpenTUI renders it", () => {
    const logo = torveLogo(120, 48, false)

    expect(logo.kind).toBe("ansi")
    expect(JSON.stringify(logo.rows)).not.toMatch(/\u001b/)
    expect(logo.rows.length).toBeLessThanOrEqual(36)
  })

  test("derives a compact mark from the supplied artwork on ordinary terminals", () => {
    expect(torveLogo(79, 30).kind).toBe("compact")
    expect(torveLogo(79, 30).rows.length).toBeLessThanOrEqual(17)
  })

  test("uses the supplied text artwork for large monochrome terminals", () => {
    expect(torveLogo(120, 60, true).kind).toBe("text")
    expect(torveLogo(120, 48, true).kind).toBe("compact")
  })

  test("keeps the prompt usable in very small terminals", () => {
    expect(torveLogo(27, 17).kind).toBe("wordmark")
  })
})
