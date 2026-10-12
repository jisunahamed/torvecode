import { cp, mkdir, rm } from "node:fs/promises"
import path from "node:path"
import pkg from "../package.json"

const dir = path.resolve(import.meta.dirname, "..")
const outdir = path.join(dir, "dist", "npm-cli")

await rm(outdir, { recursive: true, force: true })
await mkdir(path.join(outdir, "bin"), { recursive: true })
await cp(path.join(dir, "bin", "torve.cjs"), path.join(outdir, "bin", "torve.cjs"))
await Bun.write(
  path.join(outdir, "package.json"),
  JSON.stringify(
    {
      name: "@torveai/cli",
      version: pkg.version,
      description: pkg.description,
      license: pkg.license,
      repository: pkg.repository,
      bin: pkg.bin,
      files: ["bin"],
      optionalDependencies: pkg.optionalDependencies,
    },
    null,
    2,
  ) + "\n",
)
