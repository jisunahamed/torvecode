import { RGBA, TextAttributes } from "@opentui/core"
import { createMemo, For, Match, Switch } from "solid-js"
import { useTerminalDimensions } from "@opentui/solid"
import { useTheme } from "../context/theme"
import { torveLogo } from "../logo"

export function Logo(props: { compact?: boolean } = {}) {
  const theme = useTheme()
  const dimensions = useTerminalDimensions()
  const artwork = () => (props.compact ? torveLogo(40, 24) : torveLogo(dimensions().width, dimensions().height))
  const colorArtwork = createMemo(() => {
    const value = artwork()
    return value.kind === "ansi" ? value : undefined
  })
  const plainArtwork = createMemo(() => {
    const value = artwork()
    return value.kind !== "ansi" ? value : undefined
  })
  const brand = RGBA.fromHex("#A3E635")
  const monochrome = Boolean(process.env.NO_COLOR)

  return (
    <Switch>
      <Match when={colorArtwork()} keyed>
        {(value) => (
          <For each={value.rows}>
            {(row) => (
              <box flexDirection="row">
                <text selectable={false}>
                  <For each={row}>
                    {(segment) => (
                      <span
                        style={{
                          fg: segment.fg ? RGBA.fromHex(segment.fg) : theme.text.base,
                          bg: segment.bg ? RGBA.fromHex(segment.bg) : undefined,
                        }}
                      >
                        {segment.text}
                      </span>
                    )}
                  </For>
                </text>
              </box>
            )}
          </For>
        )}
      </Match>
      <Match when={plainArtwork()} keyed>
        {(value) => (
          <For each={value.rows}>
            {(line) => (
              <box flexDirection="row">
                <text
                  fg={value.kind === "text" || monochrome ? theme.text.base : brand}
                  attributes={TextAttributes.BOLD}
                  selectable={false}
                >
                  {line}
                </text>
              </box>
            )}
          </For>
        )}
      </Match>
    </Switch>
  )
}
