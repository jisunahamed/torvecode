import { Prompt, type PromptRef } from "../component/prompt"
import { createEffect, createMemo, createSignal, onMount, Show, untrack } from "solid-js"
import { useArgs } from "../context/args"
import { useRouteData } from "../context/route"
import { usePromptRef } from "../context/prompt"
import { useLocal } from "../context/local"
import { useEditorContext } from "../context/editor"
import { useData } from "../context/data"
import { useLocation } from "../context/location"
import { FormPrompt } from "./session/form"
import { Slot } from "../plugin/render"
import { useTerminalDimensions } from "@opentui/solid"
import { useTheme } from "../context/theme"
import { useUpdateNotification } from "../context/update-notification"
import { useExit } from "../context/exit"
import { FadeInText } from "../component/fade-in-text"
import { TextAttributes } from "@opentui/core"
import { Keymap } from "../context/keymap"
import { useDialog } from "../ui/dialog"
import { useClient } from "../context/client"
import { useTuiApp } from "../context/runtime"

let once = false
const placeholder = {
  normal: ["What should we build?", "Trace a bug through this codebase", "Review and improve this project"],
  shell: ["ls -la", "git status", "pwd"],
}

export function Home() {
  const route = useRouteData("home")
  const promptRef = usePromptRef()
  const [ref, setRef] = createSignal<PromptRef | undefined>()
  const args = useArgs()
  const local = useLocal()
  const editor = useEditorContext()
  const data = useData()
  const location = useLocation()
  const dimensions = useTerminalDimensions()
  const theme = useTheme()
  const client = useClient()
  const app = useTuiApp()
  const dialog = useDialog()
  const [sidebar, setSidebar] = createSignal(true)
  Keymap.createLayer(() => ({
    mode: "global",
    enabled: () => dialog.stack.length === 0,
    commands: [
      {
        id: "home.sidebar.toggle",
        bind: "ctrl+b",
        title: "Toggle sidebar",
        palette: true,
        run: () => {
          setSidebar((value) => !value)
        },
      },
    ],
  }))
  // Global MCP elicitations can arrive without a session route, so keep them reachable from Home.
  const currentLocation = () => route.location ?? data.location.default()
  const forms = createMemo(() => data.session.form.list("global", currentLocation()) ?? [])
  let sent = false

  // Track only the route location and (when absent) the default location; location.set
  // reads other signals internally and tracking them would re-assert the route location
  // after the user overrides it with /cd.
  createEffect(() => {
    const target = currentLocation()
    untrack(() => location.set(target))
  })

  onMount(() => {
    editor.clearSelection()
  })

  const bind = (r: PromptRef | undefined) => {
    setRef(r)
    promptRef.set(r)
    if (once || !r || route.prompt || !args.prompt) return
    r.set({ text: args.prompt, files: [], agents: [], pasted: [] })
    once = true
  }

  createEffect(() => {
    const composer = ref()
    const prompt = route.prompt
    if (!composer || prompt?.text === undefined) return
    untrack(() => composer.set(prompt))
  })

  // Wait for everything submit needs before auto-submitting --prompt; it runs once.
  createEffect(() => {
    const r = ref()
    if (sent) return
    if (!r) return
    if (!local.model.ready || !local.model.catalogReady) return
    if (!local.agent.current() || !local.model.current()) return
    if (!args.prompt) return
    if (r.current.text !== args.prompt) return
    sent = true
    r.submit()
  })

  return (
    <>
      <box flexGrow={1} minWidth={0} minHeight={0} flexDirection="column">
        <HomeHeader
          version={app.version}
          workspace={currentLocation().directory}
          model={local.model.current()?.modelID}
          status={client.connection.status()}
          onToggleSidebar={() => setSidebar((value) => !value)}
        />
        <box flexGrow={1} minWidth={0} minHeight={0} flexDirection="row">
          <box flexGrow={1} minWidth={0} minHeight={0} paddingLeft={2} paddingRight={2}>
            <box flexGrow={1} minHeight={0} paddingTop={1}>
              <box flexDirection="row" gap={2}>
                <text fg={theme.text.action.primary.selected} attributes={TextAttributes.BOLD} selectable={false}>
                  ›
                </text>
                <box gap={1}>
                  <text fg={theme.text.base} attributes={TextAttributes.BOLD} selectable={false}>
                    What can I help you build?
                  </text>
                  <Show when={dimensions().height >= 18}>
                    <text fg={theme.text.muted} selectable={false}>
                      Explore, edit and debug your codebase.
                    </text>
                  </Show>
                </box>
              </box>
              <Show when={dimensions().width >= 58 && dimensions().height >= 22}>
                <box
                  marginTop={2}
                  marginLeft={3}
                  paddingTop={1}
                  paddingBottom={1}
                  paddingLeft={2}
                  border
                  borderStyle="rounded"
                  borderColor={theme.border.base}
                  gap={0}
                  flexShrink={0}
                >
                  <text fg={theme.text.muted}>Here are some things you can try:</text>
                  <text fg={theme.text.action.primary.selected}>
                    /models <span style={{ fg: theme.text.muted }}>Choose a Torve AI model</span>
                  </text>
                  <text fg={theme.text.action.primary.selected}>
                    /skills <span style={{ fg: theme.text.muted }}>Browse coding skills</span>
                  </text>
                  <text fg={theme.text.action.primary.selected}>
                    /sessions <span style={{ fg: theme.text.muted }}>Open an earlier session</span>
                  </text>
                  <text fg={theme.text.action.primary.selected}>
                    /help <span style={{ fg: theme.text.muted }}>See commands and shortcuts</span>
                  </text>
                </box>
              </Show>
              <box flexGrow={1} minHeight={0} />
              <UpdateNotification width={Math.max(1, dimensions().width - 34)} />
            </box>
          </box>
          <Show when={dimensions().width >= 110 && sidebar()}>
            <HomeSidebar workspace={currentLocation().directory} connected={torveConnected(data)} />
          </Show>
        </box>
        <box
          width="100%"
          zIndex={1000}
          flexShrink={0}
          paddingLeft={2}
          paddingRight={2}
          paddingTop={1}
          paddingBottom={1}
        >
          <Prompt ref={bind} placeholders={placeholder} disabled={forms().length > 0} />
        </box>
        <box
          height={2}
          flexShrink={0}
          flexDirection="row"
          paddingLeft={2}
          paddingRight={2}
          border={["top"]}
          borderColor={theme.border.base}
        >
          <text fg={theme.text.muted}>IN 0 · OUT 0 · context 0%</text>
          <box flexGrow={1} />
          <text fg={theme.text.muted}>ctrl+? help</text>
        </box>
      </box>
      <box width="100%" flexShrink={0}>
        <Slot path="home.footer" />
      </box>
      <Show when={forms()[0]?.id} keyed>
        {(_) => {
          const form = forms()[0]
          return form ? (
            <box position="absolute" zIndex={2000} left={0} right={0} bottom={1} paddingLeft={2} paddingRight={2}>
              <box width="100%">
                <FormPrompt form={form} />
              </box>
            </box>
          ) : null
        }}
      </Show>
    </>
  )
}

function HomeHeader(props: {
  version: string
  workspace: string
  model?: string
  status: string
  onToggleSidebar: () => void
}) {
  const theme = useTheme()
  const dimensions = useTerminalDimensions()
  const connected = () => props.status === "connected"
  return (
    <box
      height={2}
      flexShrink={0}
      flexDirection="row"
      alignItems="center"
      paddingLeft={2}
      paddingRight={1}
      gap={dimensions().width < 100 ? 1 : 2}
      border={["bottom"]}
      borderColor={theme.border.base}
    >
      <text fg={theme.text.base} attributes={TextAttributes.BOLD} wrapMode="none" flexShrink={0}>
        ⬡ TorveCode
      </text>
      <Show when={dimensions().width >= 100}>
        <text fg={theme.text.muted}>v{props.version}</text>
        <text fg={theme.border.base}>│</text>
        <text fg={theme.text.base} truncate flexGrow={1} minWidth={8}>
          ▱ {props.workspace}
        </text>
        <text fg={theme.border.base}>│</text>
      </Show>
      <text fg={theme.text.base} truncate minWidth={0} maxWidth={dimensions().width < 100 ? 20 : 30}>
        ◇ {props.model ?? "Selecting model"}
      </text>
      <box flexGrow={1} minWidth={0} />
      <text
        fg={connected() ? theme.text.feedback.success.base : theme.text.feedback.warning.base}
        wrapMode="none"
        flexShrink={0}
      >
        ● {connected() ? "Connected" : "Connecting"}
      </text>
      <Show when={dimensions().width >= 110}>
        <text fg={theme.text.muted} onMouseUp={props.onToggleSidebar}>
          ctrl+b Toggle sidebar
        </text>
      </Show>
    </box>
  )
}

function HomeSidebar(props: { workspace: string; connected: boolean }) {
  const theme = useTheme()
  return (
    <box
      width={34}
      height="100%"
      flexShrink={0}
      paddingTop={1}
      paddingBottom={1}
      paddingLeft={2}
      paddingRight={2}
      backgroundColor={theme.background.base}
      border={["left"]}
      borderColor={theme.border.base}
      gap={1}
    >
      <text fg={theme.text.action.primary.selected} attributes={TextAttributes.BOLD}>
        ▱ Workspace
      </text>
      <text fg={theme.text.base} wrapMode="word">
        {props.workspace}
      </text>
      <box height={1} border={["bottom"]} borderColor={theme.border.base} />
      <text fg={theme.text.action.primary.selected}>▤ CURRENT SESSION</text>
      <text fg={theme.text.muted}>New session</text>
      <box height={1} border={["bottom"]} borderColor={theme.border.base} />
      <text fg={theme.text.action.primary.selected}>⌁ MODIFIED FILES</text>
      <text fg={theme.text.muted}>No modified files</text>
      <box height={1} border={["bottom"]} borderColor={theme.border.base} />
      <text fg={theme.text.muted}>⌁ DIAGNOSTICS</text>
      <text fg={theme.text.feedback.success.base}>● No issues</text>
      <text fg={theme.text.muted}>{props.connected ? "Torve AI connected" : "Connection required"}</text>
    </box>
  )
}

function torveConnected(data: ReturnType<typeof useData>) {
  return Boolean(
    data.location.integration.list()?.find((integration) => integration.id === "torveai")?.connections.length,
  )
}

function UpdateNotification(props: { width: number }) {
  const update = useUpdateNotification()
  const exit = useExit()
  const theme = useTheme()
  const [hovered, setHovered] = createSignal(false)
  const backdrop = () => (hovered() ? theme.background.action.primary.hovered : theme.background.base)
  createEffect(() => {
    update.notification()
    setHovered(false)
  })

  return (
    <Show when={update.notification()} keyed>
      {(state) => {
        const remote = state.source === "server" && state.remote
        return (
          <Show when={!remote || state.type === "available"}>
            <box
              flexShrink={0}
              flexDirection="row"
              justifyContent="center"
              width={props.width}
              maxWidth="100%"
              gap={1}
              backgroundColor={hovered() ? theme.background.action.primary.hovered : undefined}
              onMouseOver={() => setHovered(true)}
              onMouseOut={() => setHovered(false)}
              onMouseUp={() => {
                if (remote) return update.dismiss()
                if (state.type === "installed") return exit()
                update.open?.("notification")
              }}
            >
              <FadeInText fg={theme.text.muted} backdrop={backdrop()}>
                <Show when={!remote}>
                  <span style={{ fg: theme.text.action.primary.selected }}>
                    {state.type === "installed" ? "/exit" : "/update"}
                  </span>
                </Show>
                {remote
                  ? "remote server update available"
                  : state.type === "installed"
                    ? ` restart to use v${state.version}`
                    : ` to install v${state.version}`}
              </FadeInText>
            </box>
          </Show>
        )
      }}
    </Show>
  )
}
