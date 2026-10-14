import type { IntegrationOAuthMethodRegistration } from "@opencode/plugin/effect/integration"
import { define } from "@opencode/plugin/effect/plugin"
import { Clock, Duration, Effect, Schema, Stream } from "effect"
import { App } from "../../app.js"
import { Bus } from "../../bus.js"
import { Credential } from "../../credential.js"
import { Integration } from "../../integration.js"
import { Model } from "../../model.js"
import { Provider } from "../../provider.js"

const integrationID = Integration.ID.make("torveai")
const providerID = Provider.ID.make("torveai")
const methodID = Integration.MethodID.make("website")
const server = (process.env.TORVE_API_URL || "https://api.torveai.com").replace(/\/+$/, "")
const deviceGrant = "urn:ietf:params:oauth:grant-type:device_code"

const Device = Schema.Struct({
  device_code: Schema.String,
  user_code: Schema.String,
  verification_uri: Schema.String,
  verification_uri_complete: Schema.String,
  expires_in: Schema.Number,
  interval: Schema.Number,
})

const Token = Schema.Struct({
  access_token: Schema.String,
  refresh_token: Schema.String,
  expires_in: Schema.Number,
})

const Catalog = Schema.Struct({
  data: Schema.Array(
    Schema.Struct({
      id: Schema.String,
      display_name: Schema.NullOr(Schema.String),
      context_window: Schema.NullOr(Schema.Number),
      max_output_tokens: Schema.NullOr(Schema.Number),
      tool_use: Schema.Boolean,
    }),
  ),
})

const ErrorBody = Schema.Struct({
  error: Schema.optional(Schema.String),
  message: Schema.optional(Schema.String),
})

const website = (app: App.Info) =>
  ({
    integrationID,
    method: {
      id: methodID,
      type: "oauth",
      label: "Continue with Torve website",
    },
    authorize: () =>
      request(
        `${server}/cli/auth/device`,
        {
          method: "POST",
          headers: jsonHeaders(app),
          body: JSON.stringify({ device_name: deviceName() }),
        },
        Device,
      ).pipe(
        Effect.flatMap((device) =>
          Clock.currentTimeMillis.pipe(
            Effect.map((created) => ({
              mode: "auto" as const,
              url: device.verification_uri_complete,
              instructions: `Open ${device.verification_uri} and enter code: ${device.user_code}`,
              expiresAt: created + device.expires_in * 1000,
              callback: poll(device, app),
            })),
          ),
        ),
      ),
    refresh: (credential) =>
      request(
        `${server}/cli/auth/token`,
        {
          method: "POST",
          headers: jsonHeaders(app),
          body: JSON.stringify({ grant_type: "refresh_token", refresh_token: credential.refresh }),
        },
        Token,
      ).pipe(Effect.map((token) => oauthCredential(token))),
  }) satisfies IntegrationOAuthMethodRegistration

export const TorvePlugin = define({
  id: "torve.provider",
  effect: Effect.fn(function* (ctx) {
    const bus = yield* Bus.Service

    const load = Effect.fn("TorvePlugin.load")(function* () {
      const connection = process.env.TORVE_API_KEY
        ? ({ type: "env", name: "TORVE_API_KEY" } as const)
        : yield* ctx.integration.connection.active(integrationID)
      if (!connection) return { connection: undefined, models: [] as Model.Info[] }
      const credential = yield* ctx.integration.connection.resolve(connection)
      if (!credential || credential.type === "external") return { connection: undefined, models: [] as Model.Info[] }
      const token = credential.type === "oauth" ? credential.access : credential.key
      const catalog = yield* request(
        `${server}/cli/models`,
        { headers: { ...jsonHeaders(ctx.app), authorization: `Bearer ${token}` } },
        Catalog,
      ).pipe(
        Effect.timeout("3 seconds"),
        Effect.catch((cause) =>
          Effect.logWarning("failed to load Torve AI model catalog", { cause }).pipe(Effect.as({ data: [] })),
        ),
      )
      const seen = new Set<string>()
      const models = catalog.data.flatMap((item) => {
        const key = item.id.trim().toLowerCase()
        if (!item.tool_use || !key || seen.has(key)) return []
        seen.add(key)
        const id = Model.ID.make(item.id)
        return [
          {
            ...Model.Info.default(providerID, id),
            name: item.display_name || item.id,
            capabilities: { tools: true, input: ["text"], output: ["text"] },
            limit: {
              context: positiveInt(item.context_window, 128_000),
              output: positiveInt(item.max_output_tokens, 16_384),
            },
          },
        ]
      })
      return { connection, models }
    })

    let snapshot = yield* load().pipe(
      Effect.catch((cause) =>
        Effect.logWarning("failed to initialize Torve AI", { cause }).pipe(
          Effect.as({ connection: undefined, models: [] as Model.Info[] }),
        ),
      ),
    )

    yield* ctx.integration.transform((editor) => {
      editor.update(integrationID, (integration) => {
        integration.name = "Torve AI"
      })
      editor.method.update(website(ctx.app))
      editor.method.update({ integrationID, method: { type: "key", label: "Use Torve AI API key" } })
      editor.method.update({ integrationID, method: { type: "env", names: ["TORVE_API_KEY"] } })
    })

    yield* ctx.provider.transform((providers) => {
      if (!snapshot.connection || snapshot.models.length === 0) return
      providers.add({
        info: {
          id: providerID,
          integrationID,
          name: "Torve AI",
          activation: "enabled",
          package: "@opencode/ai/providers/openai-compatible",
          settings: { baseURL: `${server}/v1` },
        },
        models: snapshot.models,
        sourceConnection: snapshot.connection,
      })
    })

    const refresh = Effect.fn("TorvePlugin.refresh")(function* () {
      snapshot = yield* load().pipe(
        Effect.catch((cause) =>
          Effect.logWarning("failed to refresh Torve AI", { cause }).pipe(Effect.as(snapshot)),
        ),
      )
      yield* ctx.provider.reload()
    })
    yield* bus.subscribe(Credential.Event.Switched).pipe(
      Stream.filter((event) => event.data.integrationID === integrationID),
      Stream.runForEach(refresh),
      Effect.forkScoped({ startImmediately: true }),
    )
    yield* Effect.sleep(Duration.minutes(1)).pipe(Effect.andThen(refresh()), Effect.forever, Effect.forkScoped)
  }),
})

function poll(device: typeof Device.Type, app: App.Info): Effect.Effect<Credential.OAuth, unknown> {
  const loop = (interval: number): Effect.Effect<Credential.OAuth, unknown> =>
    Effect.gen(function* () {
      yield* Effect.sleep(interval)
      const response = yield* send(`${server}/cli/auth/token`, {
        method: "POST",
        headers: jsonHeaders(app),
        body: JSON.stringify({ grant_type: deviceGrant, device_code: device.device_code }),
      })
      if (response.ok) return oauthCredential(yield* decode(response, Token))
      const failure = yield* decode(response, ErrorBody).pipe(
        Effect.orElseSucceed(() => ({ error: undefined, message: undefined })),
      )
      const code = failure.error
      if (code === "authorization_pending") return yield* loop(interval)
      if (code === "slow_down") return yield* loop(interval + 5_000)
      if (code === "access_denied") return yield* Effect.fail(new Error("Torve website authorization was denied"))
      if (code === "expired_token") return yield* Effect.fail(new Error("Torve website authorization expired"))
      return yield* Effect.fail(new Error(failure.message || `Torve authorization failed (${response.status})`))
    })
  return loop(Math.max(device.interval * 1000, Duration.toMillis(Duration.seconds(1))))
}

function oauthCredential(token: typeof Token.Type) {
  return Credential.OAuth.make({
    type: "oauth",
    methodID,
    access: token.access_token,
    refresh: token.refresh_token,
    expires: Date.now() + token.expires_in * 1000,
    metadata: { server },
  })
}

function request<S extends Schema.Decoder<unknown>>(url: string, init: RequestInit, schema: S) {
  return send(url, init).pipe(
    Effect.flatMap((response) => {
      if (response.ok) return decode(response, schema)
      return Effect.promise(() => response.text()).pipe(
        Effect.flatMap((detail) =>
          Effect.fail(new Error(`Torve request failed (${response.status})${detail ? `: ${detail}` : ""}`)),
        ),
      )
    }),
  )
}

function send(url: string, init: RequestInit) {
  return Effect.tryPromise({
    try: (signal) => fetch(url, { ...init, signal }),
    catch: (cause) => cause,
  })
}

function decode<S extends Schema.Decoder<unknown>>(response: Response, schema: S) {
  return Effect.promise(() => response.json()).pipe(Effect.map(Schema.decodeUnknownSync(schema)))
}

function jsonHeaders(app: App.Info) {
  return { Accept: "application/json", "Content-Type": "application/json", "User-Agent": App.useragent(app) }
}

function deviceName() {
  return `${process.platform === "win32" ? "Windows" : process.platform} terminal`
}

function positiveInt(value: number | null, fallback: number) {
  return value && Number.isSafeInteger(value) && value > 0 ? value : fallback
}
