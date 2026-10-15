import fs from "fs"
import path from "path"
import { execSync } from "child_process"

const cliDir = path.resolve("packages/cli")
const v109Dir = path.join(cliDir, ".codex-release-v109")
const packedDir = path.join(cliDir, ".codex-release-v109-packed")

const freshWinExe = path.join(cliDir, "dist", "cli-win32-x64", "bin", "torve.exe")
const targetWinExe = path.join(v109Dir, "cli-win32-x64", "bin", "torve.exe")
fs.copyFileSync(freshWinExe, targetWinExe)
console.log("Copied fresh torve.exe to .codex-release-v109/cli-win32-x64/bin/torve.exe")

console.log("Repacking cli-win32-x64...")
execSync("npm pack", { cwd: path.join(v109Dir, "cli-win32-x64"), stdio: "inherit" })

const files = fs.readdirSync(path.join(v109Dir, "cli-win32-x64"))
for (const f of files) {
  if (f.endsWith(".tgz")) {
    fs.copyFileSync(path.join(v109Dir, "cli-win32-x64", f), path.join(packedDir, f))
    fs.unlinkSync(path.join(v109Dir, "cli-win32-x64", f))
    console.log(`Updated ${f} in packed directory.`)
  }
}

console.log("\n=== ALL PACKAGES IN PACKED DIR ===")
for (const f of fs.readdirSync(packedDir)) {
  const stat = fs.statSync(path.join(packedDir, f))
  console.log(`- ${f} (${(stat.size / 1024 / 1024).toFixed(2)} MB)`)
}
