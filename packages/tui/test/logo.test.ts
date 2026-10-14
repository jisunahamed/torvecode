import { describe, expect, test } from "bun:test"
import { torveLogo } from "../src/logo"

describe("Torve terminal logo", () => {
  test("keeps the full logo free of ANSI controls before OpenTUI renders it", () => {
    const logo = torveLogo(120, 48, false)

    expect(logo.kind).toBe("ansi")
    expect(JSON.stringify(logo.rows)).not.toMatch(/\u001b/)
    expect(logo.rows.length).toBeLessThanOrEqual(36)
  })

  test("uses a fixed compact mark on ordinary terminals", () => {
    const ordinary = torveLogo(79, 30)
    const smaller = torveLogo(40, 24)

    expect(ordinary.kind).toBe("compact")
    expect(smaller).toEqual(ordinary)
    expect(ordinary.rows.length).toBe(7)
  })

  test("uses the supplied text artwork for large monochrome terminals", () => {
    expect(torveLogo(120, 60, true).kind).toBe("text")
    expect(torveLogo(120, 48, true).kind).toBe("compact")
  })

  test("keeps the prompt usable in very small terminals", () => {
    expect(torveLogo(27, 17).kind).toBe("wordmark")
    expect(torveLogo(27, 17).rows).toEqual(["⬡ TORVECODE"])
  })
})
