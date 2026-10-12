import { RGBA, TextAttributes } from "@opentui/core"
import { For, Show } from "solid-js"
import { useTerminalDimensions } from "@opentui/solid"
import { useTheme } from "../context/theme"
import { torveLogo } from "../logo"

export function Logo() {
  const theme = useTheme()
  const dimensions = useTerminalDimensions()

  const lime = RGBA.fromHex("#A3E635")
  const compact = () => dimensions().width < 80 || dimensions().height < 38
  const torveLines = () => torveLogo(dimensions().width, dimensions().height).split("\n")

  return (
    <box>
      <Show
        when={!compact()}
        fallback={
          <box flexDirection="row" gap={1}>
            <text fg={theme.text.base} attributes={TextAttributes.BOLD} selectable={false}>
              TORVE
            </text>
            <text fg={lime} attributes={TextAttributes.BOLD} selectable={false}>
              AI
            </text>
          </box>
        }
      >
        <For each={torveLines()}>
          {(line) => (
            <box flexDirection="row">
              <text fg={lime} attributes={TextAttributes.BOLD} selectable={false}>
                {line}
              </text>
            </box>
          )}
        </For>
      </Show>
    </box>
  )
}
