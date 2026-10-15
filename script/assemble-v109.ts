import fs from "fs"
import path from "path"
import { execSync } from "child_process"

const cliDir = path.resolve("packages/cli")
const v108Dir = path.join(cliDir, ".codex-release-v108")
const v109Dir = path.join(cliDir, ".codex-release-v109")
const packedDir = path.join(cliDir, ".codex-release-v109-packed")

// 1. Create v109 directories
fs.rmSync(v109Dir, { recursive: true, force: true })
fs.rmSync(packedDir, { recursive: true, force: true })
fs.cpSync(v108Dir, v109Dir, { recursive: true })
fs.mkdirSync(packedDir, { recursive: true })

// 2. Copy the freshly built windows x64 binary
const freshWinExe = path.join(cliDir, "dist", "cli-win32-x64", "bin", "torve.exe")
const targetWinExe = path.join(v109Dir, "cli-win32-x64", "bin", "torve.exe")
fs.copyFileSync(freshWinExe, targetWinExe)
console.log("Copied freshly built torve.exe to v109/cli-win32-x64/bin/torve.exe")

// 3. Update version in all 6 platform package.json files
const platforms = [
  "cli-darwin-arm64",
  "cli-darwin-x64",
  "cli-linux-arm64",
  "cli-linux-x64",
  "cli-win32-arm64",
  "cli-win32-x64",
]

for (const plat of platforms) {
  const pkgPath = path.join(v109Dir, plat, "package.json")
  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"))
  pkg.version = "1.0.9"
  fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + "\n", "utf8")

  // Pack the platform package
  console.log(`Packing ${plat}...`)
  execSync("npm pack", { cwd: path.join(v109Dir, plat), stdio: "inherit" })
  
  // Move the .tgz file to packedDir
  const files = fs.readdirSync(path.join(v109Dir, plat))
  for (const f of files) {
    if (f.endsWith(".tgz")) {
      fs.copyFileSync(path.join(v109Dir, plat, f), path.join(packedDir, f))
      fs.unlinkSync(path.join(v109Dir, plat, f))
      console.log(`Saved ${f} to packed directory.`)
    }
  }
}

// 4. Pack main @torveai/cli package
console.log("Packing main @torveai/cli package...")
execSync("bun run script/pack-npm.ts", { cwd: cliDir, stdio: "inherit" })
const npmCliDir = path.join(cliDir, "dist", "npm-cli")
execSync("npm pack", { cwd: npmCliDir, stdio: "inherit" })
const npmCliFiles = fs.readdirSync(npmCliDir)
for (const f of npmCliFiles) {
  if (f.endsWith(".tgz")) {
    fs.copyFileSync(path.join(npmCliDir, f), path.join(packedDir, f))
    fs.unlinkSync(path.join(npmCliDir, f))
    console.log(`Saved main package ${f} to packed directory.`)
  }
}

console.log("\n=== ALL 7 PACKAGES PACKED FOR V1.0.9 ===")
const packedList = fs.readdirSync(packedDir)
for (const f of packedList) {
  const stat = fs.statSync(path.join(packedDir, f))
  console.log(`- ${f} (${(stat.size / 1024 / 1024).toFixed(2)} MB)`)
}
