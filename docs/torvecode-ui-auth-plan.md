# TorveCode UI and authentication implementation plan

## Outcome

TorveCode will open directly into a branded, responsive coding workspace. The first launch will never leave an unauthenticated user at an unusable editor: it will show a Torve connection screen with **Continue with website** and **Use API key**. After either method succeeds, the same Torve AI model catalog, quota, sessions, tools, and coding interface will be available.

This plan is based on the current TypeScript/SolidJS/OpenTUI repository. The older Go UI plan does not apply to this implementation.

## Current findings

- The current home screen renders the large logo above a generic prompt, while the target workspace composition exists only in parts: `SessionHeader`, `SessionWelcome`, `Sidebar`, and `Prompt` are separate components and do not yet form the requested first-screen experience.
- `torve-logo.ansi` and `torve-logo.txt` are bundled, but narrow terminals use a generated resampling algorithm. That algorithm distorts the logo when the terminal shrinks.
- The Torve website/backend already has `/cli/auth/device`, `/cli/auth/token`, `/cli/auth/revoke`, `/cli/models`, and `/cli/account` endpoints and an approval page at `/cli/authorize`.
- The CLI's reusable device provider still carries OpenCode Console defaults and `/auth/device/*` endpoint assumptions. Torve login is therefore not a dedicated, guaranteed first-party path.
- The CLI already filters the model picker to the `torveai` provider, removes known unsupported models, and deduplicates exact model IDs. Authentication and catalog wiring must feed that provider consistently.

## Design direction

### Visual tokens

| Role | Value |
|---|---|
| Canvas | `#0B0D0F` |
| Raised surface | `#11151A` |
| Border | `#293038` |
| Primary text | `#F4F7F5` |
| Muted text | `#94A39A` |
| Torve lime | `#A3E635` |

The memorable element is the Torve cube mark and lime focus rail. Everything else stays quiet and functional. Rounded card chrome will be limited to the composer and actionable dialogs so the terminal does not look like a web dashboard copied into a console.

### Responsive layout

Wide terminal, at least 110 columns:

```text
┌ TorveCode  v1.x │ workspace │ model │ ● Connected │ Ctrl+B ┐
│ conversation and tool timeline                  │ workspace │
│                                                  │ session   │
│                                                  │ changes   │
│                                                  │ status    │
├──────────────────────────────────────────────────┴───────────┤
│ › Ask anything…        @ files   / commands   Enter to send │
├──────────────────────────────────────────────────────────────┤
│ IN 0  OUT 0  context 0/128K                 LSP ready       │
└──────────────────────────────────────────────────────────────┘
```

Medium terminal, 72–109 columns: the right sidebar is hidden by default and opens as an overlay with `Ctrl+B`; the full header remains but low-priority labels collapse.

Narrow terminal, below 72 columns: compact header, single-column conversation, full-width composer, and a fixed Torve mark/wordmark. No generated logo scaling is used.

Short terminal, below 22 rows: artwork and welcome suggestions disappear before the composer or status bar. Conversation space always keeps a usable minimum height.

### Logo rules

- Use the ANSI logo only when its measured width and height fit without scaling.
- Use the text logo only when it fits exactly and `NO_COLOR` is active.
- Add a hand-authored compact cube plus `TorveCode` wordmark for medium terminals.
- Use `⬡ TorveCode` for narrow headers.
- Never sample, squash, or crop either source logo. Terminal resize must switch variants atomically.

## Workstream 1: compose the requested UI

1. Turn the session route into the primary product surface: compact header, conversation timeline, optional right sidebar, composer, and status strip.
2. Replace the current generic home composition with a lightweight empty-session version of that same shell so the interface does not visibly change after the first message.
3. Separate user prompts, assistant responses, thinking, tool activity, file reads, edits, shell commands, and errors with stable visual rows and clear labels.
4. Keep tool output collapsible. Show running state, retry count, elapsed time, and completion state without moving the composer.
5. Keep the composer pinned at the bottom. Preserve backspace, multiline entry, text paste, image paste, file mentions, command completion, and mouse/keyboard scrolling.
6. Make every modal and command palette escapable with `Esc` and `Ctrl+C`; add a visible Back action when a dialog has a parent step.
7. Display input/output tokens and context usage from real response metadata. Show compaction events in the timeline and update the context meter after compaction.

Primary files:

- `packages/tui/src/routes/home.tsx`
- `packages/tui/src/routes/session/index.tsx`
- `packages/tui/src/component/session-frame.tsx`
- `packages/tui/src/routes/session/sidebar.tsx`
- `packages/tui/src/component/prompt/index.tsx`
- `packages/tui/src/logo.ts`
- `packages/tui/src/component/logo.tsx`
- `packages/tui/src/theme/assets/torve.json`

## Workstream 2: dedicated Torve authentication

### Credential priority

1. `TORVE_API_KEY` from the current process.
2. Saved Torve website session.
3. Saved Torve API key.
4. No credential: open connection onboarding.

An invalid explicit environment key fails visibly and does not silently fall back. Website and saved-key sessions may refresh normally.

### First-launch behavior

After the local service is ready, TorveCode checks only local credential state first. This keeps launch fast. If no Torve credential exists, it opens a blocking but escapable connection screen:

- **Continue with website** — creates a device grant, opens `https://torveai.com/cli/authorize?...`, shows the URL and user code for SSH, polls every five seconds, stores the rotating session securely, fetches the account and model catalog, then enters the workspace.
- **Use API key** — accepts a hidden key, verifies it against `/cli/account` or `/cli/models`, stores it in the OS credential store, then enters the workspace.
- **Exit** — closes cleanly. Coding actions stay disabled until connected.

If an existing website session is expired or revoked, the UI returns to this screen with **Reconnect with website** and preserves local sessions. If the network is temporarily unavailable, it shows Retry and Exit instead of treating the user as logged out.

### CLI/backend contract

Implement a dedicated `torveai` integration instead of reusing the OpenCode Console identity:

- Base API: `https://api.torveai.com`
- Device issue: `POST /cli/auth/device`
- Token exchange/refresh: `POST /cli/auth/token`
- Revoke: `POST /cli/auth/revoke`
- Catalog: `GET /cli/models`
- Account: `GET /cli/account`
- Inference: existing Torve gateway routes
- Browser approval: use the backend-provided `verification_uri_complete`

The provider must use `torveai` as its integration/provider ID and `TORVE_API_KEY` as its environment key. No visible OpenCode labels, URLs, provider choices, or account names remain.

The website response shape and CLI decoder will be covered by contract tests so endpoint naming or token fields cannot drift again.

Primary files:

- new Torve provider under `packages/core/src/plugin/provider/`
- `packages/core/src/plugin/internal.ts`
- `packages/core/src/plugin/torve-models.ts`
- `packages/cli/src/commands/handlers/auth/login.ts`
- `packages/cli/src/commands/handlers/default.ts`
- new TUI connection dialog under `packages/tui/src/component/`
- backend CLI auth controller/service and shared contracts under `api-website/backend/`

## Workstream 3: model reliability

1. Treat `/cli/models` as the canonical catalog for website sessions and API keys.
2. Normalize aliases case-insensitively, but deduplicate by canonical alias plus protocol/operation so distinct valid routes are not accidentally removed.
3. Only expose active, text-output, tool-capable models whose protocol and operation the CLI implements.
4. Validate each selected model before the first task and provide a clear replacement when a model was removed from the account.
5. Route OpenAI-compatible Torve models through the Torve gateway regardless of their underlying upstream vendor name.
6. Preserve RPM, TPM, monthly allowance, and balance errors from the gateway as actionable messages with reset information.

## Workstream 4: startup performance

- Render the shell and local authentication state before update checks, remote catalog refresh, MCP discovery, LSP startup, repository indexing, or model health checks.
- Run those tasks after first paint and show their state in the header/sidebar.
- Reuse the background service only when its version is compatible; put restart progress on screen rather than leaving a blank terminal.
- Add timestamps for process start, service ready, first frame, credential resolved, and catalog ready.

Acceptance target: first usable frame within five seconds on the supported Windows reference machine, with a stricter internal target of two seconds when the background service is already running.

## Implementation sequence

1. Add UI stories for wide, medium, narrow, short, disconnected, connecting, connected, streaming, tool-running, error, and compacted-session states.
2. Replace distorted logo scaling with fixed responsive variants.
3. Build the common workspace shell and apply it to both home and session routes.
4. Add the Torve integration and connect it to existing website endpoints.
5. Add startup authentication gating and reconnect behavior.
6. Normalize and verify the Torve model catalog.
7. Move non-critical startup work after first paint and measure launch timing.
8. Run focused tests, full checks, Windows/macOS/Linux builds, then perform manual terminal-size and auth-flow tests.
9. Commit and push the CLI and website/backend changes. Publish npm only after the full acceptance matrix passes.

## Acceptance matrix

- At 50, 72, 90, 110, and 160 columns the logo is readable, never stretched, and the composer remains usable.
- At 16, 22, and 35 rows the header, conversation, composer, and status strip do not overlap.
- A signed-out first launch immediately offers website login and API-key login.
- Website approval connects the CLI without copying an API key and survives restart through secure storage.
- API-key login works without a browser; invalid keys do not fall back silently.
- Revoking a device in the website prevents the next inference or refresh and triggers reconnect UI.
- The model list contains only unique, active, working Torve models; selection and tool calls work for every listed model.
- User/assistant messages, thinking, reads, edits, shell commands, retries, token usage, compaction, and errors render and scroll correctly.
- Every dialog can return with `Esc`/`Ctrl+C`; no empty command palette traps the user.
- A warm launch shows a usable frame in under two seconds; a cold launch stays under five seconds.
- `NO_COLOR`, light terminals, Windows Terminal, PowerShell, CMD, macOS Terminal, and common Linux terminals remain legible.

## Release rule

Do not publish the next npm version immediately after visual changes. First create local platform builds, install the Windows package from its packed tarball, complete both authentication paths against staging/production as appropriate, run a real coding task with file read/edit/tool calls, and capture wide/narrow screenshots. Publish all six platform packages and `@torveai/cli` only after those checks pass.
