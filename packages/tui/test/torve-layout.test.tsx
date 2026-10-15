import { expect, test } from "bun:test"
import { createTestRenderer } from "@opentui/core/testing"
import { Effect, FileSystem } from "effect"
import { Global } from "@opencode/util/global"
import { createEventStream, createFetch, json, directory } from "./fixture/tui-client"
import { tmpdir } from "./fixture/fixture"

for (const chat of [true, false])
  for (const [width, height] of [
    [140, 40],
    [80, 30],
    [44, 20],
  ]) {
    test(`Torve ${chat ? "chat" : "home"} fits ${width}x${height}`, async () => {
      await using state = await tmpdir()
      const setup = await createTestRenderer({ width, height, useThread: false, kittyKeyboard: true })
      setup.renderer.start()
      const events = createEventStream()
      const session = {
        id: "ses_test",
        title: "Simple Hello",
        projectID: "project",
        location: { directory },
        cost: 0,
        tokens: { input: 0, output: 0, reasoning: 0, cache: { read: 0, write: 0 } },
        time: { created: 0, updated: 0 },
      }
      const calls = createFetch((url) => {
        if (url.pathname === "/api/session") return json({ data: chat ? [session] : [], cursor: {} })
        if (url.pathname === "/api/session/ses_test") return json({ data: session })
        if (url.pathname === "/api/session/ses_test/message")
          return json({
            data: [
              {
                id: "msg_a",
                type: "assistant",
                agent: "build",
                model: { providerID: "demo", id: "demo-model" },
                content: [{ type: "text", text: "Ready to work on your codebase." }],
                finish: "stop",
                time: { created: 2, completed: 3 },
              },
              { id: "msg_u", type: "user", text: "hi", time: { created: 1 } },
            ],
            cursor: {},
          })
        if (url.pathname === "/api/session/ses_test/background") return json({ data: [] })
        if (
          url.pathname === "/api/session/ses_test/inbox" ||
          url.pathname === "/api/session/ses_test/permission" ||
          url.pathname === "/api/session/ses_test/diff"
        )
          return json({ data: [] })
      }, events)
      const server = Bun.serve({ port: 0, idleTimeout: 0, fetch: (request) => calls.fetch(request) })
      const { run } = await import("../src/app")
      const task = Effect.runPromise(
        run({
          app: { name: "TorveCode", version: "1.0.5", channel: "test" },
          server: { endpoint: { url: server.url.toString() } },
          config: {
            get: async () => ({
              animations: false,
              tabs: { mode: "off" },
              theme: { name: "torve" },
              keybinds: { "session.sidebar.toggle": "ctrl+b" },
            }),
            update: async () => ({}),
          },
          packages: { prepare: async () => ({ directory: "" }) },
          terminalHandoff: async () => ({ renderer: setup.renderer, mode: "dark", complete: () => {} }),
          args: chat ? { sessionID: "ses_test" } : {},
          log: () => {},
        }).pipe(Effect.provide(Global.layerWith({ state: state.path })), Effect.provide(FileSystem.layerNoop({}))),
      )
      try {
        await setup.waitForFrame((frame) => frame.includes(chat ? "hi" : "Ask anything"))
        await setup.waitForVisualIdle()
        const frame = setup.captureCharFrame()
        expect(frame).toContain("TorveCode")
        if (!chat) expect(frame).toContain("IN 0")
        expect(frame).toContain(chat ? "Build" : "What can I help you build?")
        if (width >= 110) {
          expect(frame).toContain("CURRENT SESSION")
          if (!chat) expect(frame).toContain("DIAGNOSTICS")
          expect(frame).toContain("ctrl+b")
        } else expect(frame).not.toContain("CURRENT SESSION")
      } finally {
        if (!setup.renderer.isDestroyed) setup.renderer.destroy()
        await task.finally(() => server.stop(true))
      }
    }, 20_000)
  }
