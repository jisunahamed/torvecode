# Maintenance and privacy

The upstream repository is archived. Review its successor and Go dependency advisories regularly, but port changes manually against the pinned provenance commit. Preserve the MIT license and upstream notices.

Prompts and responses are stored only in the local Torvecode session database; the Torve gateway retains request metadata according to its published policy and does not persist prompt content. Logs must redact Authorization, x-api-key, device codes and refresh tokens. Release checks scan source and artifacts for secrets.

Revoking a dashboard device prevents new requests. Uninstalling the binary does not delete local `.torvecode` history or revoke an account API key; users control those actions separately.
