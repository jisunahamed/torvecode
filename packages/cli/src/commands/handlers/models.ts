import { OpenCode } from "@opencode/client"
import { Service } from "@opencode/client/effect/service"
import { Effect, Option } from "effect"
import { EOL } from "node:os"
import { Commands } from "../commands"
import { Runtime } from "../../framework/runtime"
import { ServerConnection } from "../../services/server-connection"

export default Runtime.handler(
  Commands.commands.models,
  Effect.fn("cli.models")(function* (input) {
    const server = yield* ServerConnection.resolve({
      server: Option.getOrUndefined(input.server),
      standalone: input.standalone,
      mismatch: "replace",
    })
    const client = OpenCode.make({
      baseUrl: server.endpoint.url,
      headers: Service.headers(server.endpoint),
    })
    const response = yield* Effect.promise(() => loadModels(client))
    const models = response.data
      .map((model) => `${model.providerID}/${model.id}`)
      .toSorted((a, b) => a.localeCompare(b))
    if (models.length > 0) process.stdout.write(models.join(EOL) + EOL)
  }),
)

function loadModels(
  client: ReturnType<typeof OpenCode.make>,
  attempt = 0,
): Promise<Awaited<ReturnType<typeof client.model.list>>> {
  return client.model.list({ location: { directory: process.cwd() } }).catch(async (cause) => {
    if (attempt >= 9 || !pluginActivationPending(cause)) throw cause
    await Bun.sleep(Math.min(1_000, 100 * 2 ** attempt))
    return loadModels(client, attempt + 1)
  })
}

function pluginActivationPending(cause: unknown) {
  return cause instanceof Error && cause.message.includes("plugins did not finish activating")
}
