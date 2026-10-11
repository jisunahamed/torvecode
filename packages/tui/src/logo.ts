import txtLogo from "../assets/torve-logo.txt?raw"
import ansiLogo from "../assets/torve-logo.ansi?raw"

// Torve cube robot — responsive: wide → ANSI truecolor, narrow or NO_COLOR → plain ASCII
export const torve = {
  txt: txtLogo,
  ansi: ansiLogo,
}

function useAnsi(): boolean {
  // NO_COLOR env disables ANSI for accessibility / light terminal
  if (typeof process !== "undefined" && process.env?.NO_COLOR) return false
  return true
}

export function torveLogo(width: number): string {
  if (width < 12) return "TORVECODE"
  if (width < 80 || !useAnsi()) return torve.txt
  return torve.ansi
}

// Keep opencode logo for fallback / tests
export const logo = {
  left: ["                   ", "█▀▀█ █▀▀█ █▀▀█ █▀▀▄", "█__█ █__█ █^^^ █__█", "▀▀▀▀ █▀▀▀ ▀▀▀▀ ▀~~▀"],
  right: ["             ▄     ", "█▀▀▀ █▀▀█ █▀▀█ █▀▀█", "█___ █__█ █__█ █^^^", "▀▀▀▀ ▀▀▀▀ ▀▀▀▀ ▀▀▀▀"],
}

export const go = {
  left: ["    ", "█▀▀▀", "█_^█", "▀▀▀▀"],
  right: ["    ", "█▀▀█", "█__█", "▀▀▀▀"],
}
