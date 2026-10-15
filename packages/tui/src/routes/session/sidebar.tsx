import { useData } from "../../context/data"
import { createEffect, createMemo, createResource, For, Show } from "solid-js"
import { useTheme } from "../../context/theme"
import { useConfig } from "../../config"
import { Slot } from "../../plugin/render"
import { withTimestampedFallback } from "@opencode/util/session-title-fallback"
import { TextAttributes } from "@opentui/core"
import "../../component/title-shimmer"
import { FilePath } from "../../ui/file-path"
import { useClient } from "../../context/client"

import { getScrollAcceleration } from "../../util/scroll"
import { SESSION_SIDEBAR_WIDTH } from "../../ui/layout"

export function Sidebar(props: { sessionID: string }) {
  const data = useData()
  const theme = useTheme()
  const config = useConfig().data
  const client = useClient()
  const session = createMemo(() => data.session.get(props.sessionID))
  const scrollAcceleration = createMemo(() => getScrollAcceleration(config))
  const [files, { refetch }] = createResource(
    () => props.sessionID,
    async (sessionID) => {
      try {
        return await client.api.session.diff({ sessionID, context: 0 })
      } catch {
        return []
      }
    },
  )
  createEffect((previous) => {
    const status = data.session.status(props.sessionID)
    if (previous === "running" && status === "idle") void refetch()
    return status
  })

  return (
    <Show when={session()}>
      <box
        backgroundColor={theme.background.base}
        width={SESSION_SIDEBAR_WIDTH}
        height="100%"
        paddingTop={1}
        paddingBottom={1}
        paddingLeft={2}
        paddingRight={2}
        position="relative"
        border={["left"]}
        borderColor={theme.border.base}
      >
        <box flexShrink={0} paddingBottom={1} gap={1} border={["bottom"]} borderColor={theme.border.base}>
          <text fg={theme.text.action.primary.selected} attributes={TextAttributes.BOLD}>
            ▱ Workspace
          </text>
          <FilePath value={session()!.location.directory} maxWidth={SESSION_SIDEBAR_WIDTH - 5} fg={theme.text.base} />
        </box>
        <box
          flexShrink={0}
          paddingTop={1}
          paddingBottom={1}
          gap={1}
          border={["bottom"]}
          borderColor={theme.border.base}
        >
          <text fg={theme.text.action.primary.selected}>▤ CURRENT SESSION</text>
          <title_shimmer
            fg={theme.text.base}
            rename={{
              pending: data.session.title.pending(props.sessionID),
              title: withTimestampedFallback(session()),
            }}
            enabled={config.animations ?? true}
            backdrop={theme.background.raised.base}
            attributes={
              data.session.title.pending(props.sessionID) && config.animations === false
                ? TextAttributes.DIM
                : TextAttributes.BOLD
            }
          >
            {withTimestampedFallback(session())}
          </title_shimmer>
        </box>
        <scrollbox
          flexGrow={1}
          minHeight={0}
          scrollAcceleration={scrollAcceleration()}
          // The sidebar only scrolls vertically; a horizontal bar steals a row during initial layout.
          horizontalScrollbarOptions={{ visible: false }}
          verticalScrollbarOptions={{
            // Use the content's reserved right padding instead of changing its width when the bar toggles.
            position: "absolute",
            right: 0,
            top: 0,
            width: 1,
            height: "100%",
            trackOptions: {
              backgroundColor: theme.background.raised.base,
              foregroundColor: theme.scrollbar.base,
            },
          }}
        >
          <box flexShrink={0} gap={1} paddingTop={1} paddingRight={1}>
            <box gap={1} paddingBottom={1} border={["bottom"]} borderColor={theme.border.base}>
              <text fg={theme.text.action.primary.selected}>⌁ MODIFIED FILES</text>
              <Show
                when={!files.loading && (files()?.length ?? 0) > 0}
                fallback={
                  <text fg={files.error ? theme.text.feedback.warning.base : theme.text.muted}>
                    {files.loading ? "Checking workspace…" : files.error ? "Changes unavailable" : "No modified files"}
                  </text>
                }
              >
                <For each={files()?.slice(0, 6)}>
                  {(file) => <FilePath value={file.file} maxWidth={SESSION_SIDEBAR_WIDTH - 6} fg={theme.text.base} />}
                </For>
                <Show when={(files()?.length ?? 0) > 6}>
                  <text fg={theme.text.muted}>+{(files()?.length ?? 0) - 6} more</text>
                </Show>
              </Show>
            </box>
            <box gap={1} paddingBottom={1} border={["bottom"]} borderColor={theme.border.base}>
              <text fg={theme.text.muted}>⌁ DIAGNOSTICS</text>
              <text fg={theme.text.feedback.success.base}>● No issues</text>
              <text fg={theme.text.muted}>LSP ready</text>
            </box>
            <Slot path="sidebar.content" input={{ sessionID: props.sessionID }} />
          </box>
        </scrollbox>

        <box flexShrink={0} gap={1} paddingTop={1}>
          <Slot path="sidebar.footer" input={{ sessionID: props.sessionID }} />
        </box>
      </box>
    </Show>
  )
}
