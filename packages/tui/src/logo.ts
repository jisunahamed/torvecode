import ansiLogo from "./assets/torve-logo.ansi?raw"

// OpenTUI's text nodes are not an ANSI parser. Strip terminal control codes before
// rendering the logo so escape bytes never enter native TextBuffers.
const ansiControls = /\u001b\[[0-?]*[ -/]*[@-~]/g
const rows = ansiLogo.replace(ansiControls, "").replaceAll("\r", "").split("\n")
const firstRow = rows.findIndex((row) => row.trim().length > 0)
const lastRow = rows.findLastIndex((row) => row.trim().length > 0)
const wideLogo = rows.slice(firstRow, lastRow + 1).map((row) => row.trimEnd()).join("\n")

// Large terminals get the supplied Torve artwork; smaller screens keep a compact
// wordmark so the home prompt stays visible without scrolling.
export function torveLogo(width: number, height: number): string {
  if (width < 80 || height < 38) return "TORVE AI"
  return wideLogo
}

// Compact Torve wordmarks for scrollback and mini-mode splash screens.
export const logo = {
  left: ["                   ", "  /\\               ", " |◉ ◉|              ", "  \\_/               "],
  right: ["        ", "TORVE   ", "AI      ", "        "],
}

export const go = {
  left: ["   ", " /\\", "|◉◉|", " \\_/"],
  right: ["", "TORVE", " AI", ""],
}
