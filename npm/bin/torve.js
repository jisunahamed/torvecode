#!/usr/bin/env node
'use strict';

const { spawnSync } = require('node:child_process');

const key = `${process.platform}-${process.arch}`;
const packages = {
  'darwin-arm64': '@torveai/cli-darwin-arm64',
  'darwin-x64': '@torveai/cli-darwin-x64',
  'linux-arm64': '@torveai/cli-linux-arm64',
  'linux-x64': '@torveai/cli-linux-x64',
  'win32-arm64': '@torveai/cli-win32-arm64',
  'win32-x64': '@torveai/cli-win32-x64'
};
const packageName = packages[key];
if (!packageName) {
  console.error(`Torvecode does not support ${process.platform}/${process.arch}. Supported: Windows, macOS and Linux on x64 or ARM64.`);
  process.exit(1);
}

let executable;
try {
  executable = require.resolve(`${packageName}/bin/${process.platform === 'win32' ? 'torve.exe' : 'torve'}`);
} catch {
  console.error(`The Torvecode binary package ${packageName} is missing. Reinstall @torveai/cli without disabling optional dependencies.`);
  process.exit(1);
}

const result = spawnSync(executable, process.argv.slice(2), { stdio: 'inherit', windowsHide: false });
if (result.error) { console.error(`Could not start Torvecode: ${result.error.message}`); process.exit(1); }
if (result.signal) { process.kill(process.pid, result.signal); }
process.exit(result.status == null ? 1 : result.status);
