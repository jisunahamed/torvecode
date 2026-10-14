import ansiLogo from "./assets/torve-logo.ansi" with { type: "text" }
import textLogo from "./assets/torve-logo.txt" with { type: "text" }

export type LogoSegment = {
  text: string
  fg?: string
  bg?: string
}

const ansiControl = /\u001b\[[0-?]*[ -/]*[@-~]/g
const ansiSgr = /\u001b\[([0-9;]*)m/g

const trim = (value: string) => {
  const rows = value.replaceAll("\r", "").split("\n")
  const first = rows.findIndex((row) => row.trim().length > 0)
  const last = rows.findLastIndex((row) => row.trim().length > 0)
  return rows.slice(first, last + 1).map((row) => row.trimEnd())
}

const plainAnsiRows = trim(ansiLogo.replace(ansiControl, ""))
const plainTextRows = trim(textLogo)

function parseAnsi() {
  const rows: LogoSegment[][] = [[]]
  const state: { fg?: string; bg?: string; inverse: boolean } = { inverse: false }
  let cursor = 0

  function append(text: string) {
    for (const [index, part] of text.replace(ansiControl, "").replaceAll("\r", "").split("\n").entries()) {
      if (index > 0) rows.push([])
      if (!part) continue
      const fg = state.inverse ? state.bg : state.fg
      const bg = state.inverse ? state.fg : state.bg
      const current = rows.at(-1)!
      const previous = current.at(-1)
      if (previous && previous.fg === fg && previous.bg === bg) {
        previous.text += part
        continue
      }
      current.push({ text: part, fg, bg })
    }
  }

  for (const match of ansiLogo.matchAll(ansiSgr)) {
    append(ansiLogo.slice(cursor, match.index))
    const codes = (match[1] || "0").split(";").map(Number)
    for (let index = 0; index < codes.length; index++) {
      const code = codes[index]
      if (code === 0) Object.assign(state, { fg: undefined, bg: undefined, inverse: false })
      if (code === 7) state.inverse = true
      if (code === 27) state.inverse = false
      if ((code === 38 || code === 48) && codes[index + 1] === 2) {
        const color = `#${codes
          .slice(index + 2, index + 5)
          .map((value) => value.toString(16).padStart(2, "0"))
          .join("")}`
        if (code === 38) state.fg = color
        if (code === 48) state.bg = color
        index += 4
      }
    }
    cursor = match.index + match[0].length
  }
  append(ansiLogo.slice(cursor).replace(ansiControl, ""))

  const first = rows.findIndex((row) => row.some((segment) => segment.text.trim().length > 0))
  const last = rows.findLastIndex((row) => row.some((segment) => segment.text.trim().length > 0))
  return rows.slice(first, last + 1)
}

const colorRows = parseAnsi()

function compact(rows: string[], width: number, height: number) {
  const sourceWidth = Math.max(...rows.map((row) => row.length))
  const targetWidth = Math.max(22, Math.min(sourceWidth, width - 4))
  const targetHeight = Math.max(4, Math.min(15, height - 13))
  const sampled = Array.from({ length: targetWidth }, (_, x) => Math.floor((x * sourceWidth) / targetWidth))
  return Array.from({ length: targetHeight }, (_, y) => {
    const source = Math.floor((y * rows.length) / targetHeight)
    const next = Math.min(rows.length - 1, source + 1)
    const top = rows[source] ?? ""
    const bottom = rows[next] ?? ""
    return (
      sampled
        .map((x) => {
          const upper = (top[x] ?? " ") !== " "
          const lower = (bottom[x] ?? " ") !== " "
          if (upper && lower) return "█"
          if (upper) return "▀"
          if (lower) return "▄"
          return " "
        })
        .join("")
        .trimEnd()
    )
  })
}

export function torveLogo(width: number, height: number, noColor = Boolean(process.env.NO_COLOR)) {
  if (width >= 104 && height >= 52 && noColor) return { kind: "text" as const, rows: plainTextRows }
  if (width >= 74 && height >= 35 && !noColor) return { kind: "ansi" as const, rows: colorRows }
  if (width >= 28 && height >= 18) return { kind: "compact" as const, rows: compact(plainAnsiRows, width, height) }
  return { kind: "wordmark" as const, rows: ["TORVE AI"] }
}

export const logo = {
  left: ["                   ", "  /\\               ", " |◉ ◉|              ", "  \\_/               "],
  right: ["        ", "TORVE   ", "AI      ", "        "],
}

export const go = {
  left: ["   ", " /\\", "|◉◉|", " \\_/"],
  right: ["", "TORVE", " AI", ""],
}
