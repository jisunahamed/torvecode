import fs from "fs"
import path from "path"
import zlib from "zlib"
import { execSync } from "child_process"

const cliDir = path.resolve("packages/cli")
const v109Dir = path.join(cliDir, ".codex-release-v109")
const v1011Dir = path.join(cliDir, ".codex-release-v1011")
const packedDir = path.join(cliDir, ".codex-release-v1011-packed")

// Helper: Ensure 0755 (-rwxr-xr-x) mode for binary entries inside tarball
function fixTarMode(tarBuf: Buffer) {
  let offset = 0
  while (offset < tarBuf.length - 512) {
    if (tarBuf[offset] === 0 && tarBuf[offset + 1] === 0) break
    const name = tarBuf.toString("utf8", offset, offset + 100).replace(/\0.*$/, "")
    if (name.includes("bin/torve") && !name.endsWith(".exe")) {
      console.log(`Setting 0755 mode for ${name} in tarball`)
      // Write 000755 \0 at offset + 100 (8 bytes)
      tarBuf.write("000755 \0", offset + 100, 8, "ascii")

      // Recalculate checksum
      for (let i = 0; i < 8; i++) tarBuf[offset + 148 + i] = 0x20
      let sum = 0
      for (let i = 0; i < 512; i++) sum += tarBuf[offset + i]
      const chkStr = sum.toString(8).padStart(6, "0") + "\0 "
      tarBuf.write(chkStr, offset + 148, 8, "ascii")
    }

    const sizeStr = tarBuf.toString("utf8", offset + 124, offset + 136).replace(/\0.*$/, "").trim()
    const size = parseInt(sizeStr, 8)
    offset += 512 + Math.ceil(size / 512) * 512
  }
  return tarBuf
}

// 1. Create v1011 directories
fs.rmSync(v1011Dir, { recursive: true, force: true })
fs.rmSync(packedDir, { recursive: true, force: true })
fs.cpSync(v109Dir, v1011Dir, { recursive: true })
fs.mkdirSync(packedDir, { recursive: true })

// 2. Copy the freshly built windows x64 binary
const freshWinExe = path.join(cliDir, "dist", "cli-win32-x64", "bin", "torve.exe")
const targetWinExe = path.join(v1011Dir, "cli-win32-x64", "bin", "torve.exe")
if (fs.existsSync(freshWinExe)) {
  fs.copyFileSync(freshWinExe, targetWinExe)
  console.log("Copied freshly built torve.exe to v1011/cli-win32-x64/bin/torve.exe")
} else {
  console.warn("Warning: Freshly built torve.exe not found at", freshWinExe)
}

// 3. Update version and bin field in all 6 platform package.json files
const platforms = [
  "cli-darwin-arm64",
  "cli-darwin-x64",
  "cli-linux-arm64",
  "cli-linux-x64",
  "cli-win32-arm64",
  "cli-win32-x64",
]

for (const plat of platforms) {
  const pkgPath = path.join(v1011Dir, plat, "package.json")
  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"))
  pkg.version = "1.0.11"
  pkg.bin = {
    torve: plat.includes("win32") ? "bin/torve.exe" : "bin/torve",
  }
  fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + "\n", "utf8")

  // Pack the platform package
  console.log(`Packing ${plat}...`)
  execSync("npm pack", { cwd: path.join(v1011Dir, plat), stdio: "inherit" })

  // Move the .tgz file to packedDir and fix POSIX permissions
  const files = fs.readdirSync(path.join(v1011Dir, plat))
  for (const f of files) {
    if (f.endsWith(".tgz")) {
      const srcFile = path.join(v1011Dir, plat, f)
      const destFile = path.join(packedDir, f)
      
      // Fix tar mode for non-Windows platforms
      if (!plat.includes("win32")) {
        const rawTgz = fs.readFileSync(srcFile)
        const unzipped = zlib.gunzipSync(rawTgz)
        fixTarMode(unzipped)
        const fixedTgz = zlib.gzipSync(unzipped)
        fs.writeFileSync(destFile, fixedTgz)
        console.log(`Saved and mode-fixed ${f} in packed directory.`)
      } else {
        fs.copyFileSync(srcFile, destFile)
        console.log(`Saved ${f} to packed directory.`)
      }
      fs.unlinkSync(srcFile)
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

console.log("\n=== ALL 7 PACKAGES PACKED FOR V1.0.11 ===")
const packedList = fs.readdirSync(packedDir)
for (const f of packedList) {
  const stat = fs.statSync(path.join(packedDir, f))
  console.log(`- ${f} (${(stat.size / 1024 / 1024).toFixed(2)} MB)`)
}
