import { RGBA, TextAttributes } from "@opentui/core"
import { For, type JSX } from "solid-js"
import { useTerminalDimensions } from "@opentui/solid"
import { useTheme } from "../context/theme"
import { tint } from "../theme/color"
import { go, logo, torveLogo } from "../logo"

export function Logo() {
  const theme = useTheme()
  const dimensions = useTerminalDimensions()

  // Torve palette: lime #A3E635 on #0B0D0F, not opencode amber
  const lime = RGBA.fromHex("#A3E635")
  const limeDim = RGBA.fromHex("#86C530")
  const muted = RGBA.fromHex("#77818C")

  const renderLine = (line: string, fg: RGBA, bold: boolean): JSX.Element[] => {
    const shadow = tint(theme.background.base, fg, 0.25)
    const attrs = bold ? TextAttributes.BOLD : undefined
    return Array.from(line).map((char) => {
      if (char === "_") {
        return (
          <text fg={fg} bg={shadow} attributes={attrs} selectable={false}>
            {" "}
          </text>
        )
      }
      if (char === "^") {
        return (
          <text fg={fg} bg={shadow} attributes={attrs} selectable={false}>
            ▀
          </text>
        )
      }
      if (char === "~") {
        return (
          <text fg={shadow} attributes={attrs} selectable={false}>
            ▀
          </text>
        )
      }
      if (char === ",") {
        return (
          <text fg={shadow} attributes={attrs} selectable={false}>
            ▄
          </text>
        )
      }
      return (
        <text fg={fg} attributes={attrs} selectable={false}>
          {char}
        </text>
      )
    })
  }

  // Torve cube: wide terminal gets the branded cube, narrow gets plain fallback
  const isWide = () => dimensions().width >= 80
  const torveLines = () => torveLogo(dimensions().width).split("\n")

  return (
    <box>
      {dimensions().height < 12 ? null : dimensions().width < 22 ? (
        <For each={go.right.slice(1)}>
          {(line) => <box flexDirection="row">{renderLine(line, theme.text.base, true)}</box>}
        </For>
      ) : isWide() ? (
        // Torve branded: ANSI cube (split lines rendered as-is with lime tint)
        <For each={torveLines()}>
          {(line) => <box flexDirection="row">{renderLine(line, lime, true)}</box>}
        </For>
      ) : dimensions().width < 44 ? (
        <>
          <For each={logo.left.slice(1)}>
            {(line) => <box flexDirection="row">{renderLine(line, muted, false)}</box>}
          </For>
          <For each={logo.right}>
            {(line) => <box flexDirection="row">{renderLine(line, theme.text.base, true)}</box>}
          </For>
        </>
      ) : (
        <For each={logo.left}>
          {(line, index) => (
            <box flexDirection="row" gap={1}>
              <box flexDirection="row">{renderLine(line, muted, false)}</box>
              <box flexDirection="row">{renderLine(logo.right[index()], lime, true)}</box>
            </box>
          )}
        </For>
      )}
    </box>
  )
}
