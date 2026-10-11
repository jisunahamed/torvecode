#!/usr/bin/env node
'use strict';

const { spawn, spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const readline = require('node:readline/promises');

const packageVersion = require('../package.json').version;

function npmInvocation(args) {
  if (process.platform !== 'win32') return { command: 'npm', args };

  const candidates = [
    process.env.npm_execpath,
    process.env.APPDATA && path.join(process.env.APPDATA, 'npm', 'node_modules', 'npm', 'bin', 'npm-cli.js'),
    path.join(path.dirname(process.execPath), 'node_modules', 'npm', 'bin', 'npm-cli.js')
  ].filter(Boolean);
  const cli = candidates.find((candidate) => fs.existsSync(candidate));
  if (!cli) throw new Error('Could not locate npm-cli.js. Reinstall Node.js/npm and try again.');
  return { command: process.execPath, args: [cli, ...args] };
}

function runNpm(args, options) {
  const invocation = npmInvocation(args);
  return spawnSync(invocation.command, invocation.args, options);
}

function parseUpdateArgs(args) {
  const options = { check: false, help: false, yes: false, target: 'beta' };
  for (let i = 0; i < args.length; i += 1) {
    switch (args[i]) {
      case '--check':
        options.check = true;
        break;
      case '--yes':
      case '-y':
        options.yes = true;
        break;
      case '--help':
      case '-h':
        options.help = true;
        break;
      case '--target':
        i += 1;
        if (!args[i]) throw new Error('--target requires a version or channel');
        options.target = args[i];
        break;
      default:
        throw new Error(`Unknown update option: ${args[i]}`);
    }
  }
  return options;
}

function showUpdateHelp() {
  console.log(`Usage: torve update [options]

Review an available npm release, edit the target, update or cancel.

Options:
  --check             Show the available version without installing
  --target <version>  Start with a channel or exact version (default: beta)
  -y, --yes           Install without the interactive choice
  -h, --help          Show this help`);
}

function validateTarget(target) {
  const value = target.trim();
  if (!/^[0-9A-Za-z][0-9A-Za-z._+-]*$/.test(value)) {
    throw new Error('Version/channel may contain only letters, numbers, dot, underscore, plus and hyphen.');
  }
  return value;
}

function resolveVersion(target) {
  const result = runNpm(
    ['view', `@torveai/cli@${target}`, 'version', '--json'],
    { encoding: 'utf8', windowsHide: true }
  );
  if (result.error) throw result.error;
  if (result.status !== 0) {
    const detail = (result.stderr || result.stdout || '').trim();
    throw new Error(detail || `Could not find @torveai/cli@${target}`);
  }
  const parsed = JSON.parse(result.stdout);
  return Array.isArray(parsed) ? parsed.at(-1) : parsed;
}

function showUpdate(current, target, available) {
  console.log(`Current:   ${current}`);
  console.log(`Target:    ${target}`);
  console.log(`Available: ${available}`);
  console.log(`Changes:   https://github.com/jisunahamed/torvecode/releases/tag/v${available}`);
}

async function handleUpdate(args) {
  const options = parseUpdateArgs(args);
  if (options.help) {
    showUpdateHelp();
    return 0;
  }
  let target = validateTarget(options.target);
  let available = resolveVersion(target);

  showUpdate(packageVersion, target, available);
  if (options.check) return 0;

  if (available === packageVersion) {
    console.log('Torvecode is already up to date.');
    return 0;
  }

  if (!options.yes) {
    if (!process.stdin.isTTY || !process.stdout.isTTY) {
      throw new Error('Interactive confirmation requires a terminal. Use --yes after reviewing the changes.');
    }
    const prompt = readline.createInterface({ input: process.stdin, output: process.stdout });
    try {
      while (true) {
        const action = (await prompt.question('Choose [u] update, [e] edit target, [c] cancel: ')).trim().toLowerCase();
        if (action === 'c' || action === 'cancel') {
          console.log('Update cancelled.');
          return 0;
        }
        if (action === 'e' || action === 'edit') {
          const edited = await prompt.question(`Version or channel [${target}]: `);
          if (edited.trim()) target = validateTarget(edited);
          available = resolveVersion(target);
          console.log('');
          showUpdate(packageVersion, target, available);
          if (available === packageVersion) {
            console.log('Torvecode is already up to date.');
            return 0;
          }
          continue;
        }
        if (action === 'u' || action === 'update' || action === '') break;
        console.log('Enter u, e or c.');
      }
    } finally {
      prompt.close();
    }
  }

  console.log(`Updating Torvecode to ${available}...`);
  const result = runNpm(
    ['install', '-g', `@torveai/cli@${target}`, '--prefer-online'],
    { stdio: 'inherit', windowsHide: false }
  );
  if (result.error) throw result.error;
  if (result.status !== 0) {
    if (process.platform === 'win32') {
      console.error('Update could not replace torve.exe. Close any other Torvecode windows, then run `torve update` again.');
    }
    return result.status == null ? 1 : result.status;
  }
  console.log(`Torvecode ${available} installed. Start a new session with \`torve\`.`);
  return 0;
}

const key = `${process.platform}-${process.arch}`;
const packages = {
  'darwin-arm64': '@torveai/cli-darwin-arm64',
  'darwin-x64': '@torveai/cli-darwin-x64',
  'linux-arm64': '@torveai/cli-linux-arm64',
  'linux-x64': '@torveai/cli-linux-x64',
  'win32-arm64': '@torveai/cli-win32-arm64',
  'win32-x64': '@torveai/cli-win32-x64'
};
async function main() {
  const args = process.argv.slice(2);
  if (args[0] === 'update') return handleUpdate(args.slice(1));

  const packageName = packages[key];
  if (!packageName) {
    console.error(`Torvecode does not support ${process.platform}/${process.arch}. Supported: Windows, macOS and Linux on x64 or ARM64.`);
    return 1;
  }

  let executable;
  try {
    executable = require.resolve(`${packageName}/bin/${process.platform === 'win32' ? 'torve.exe' : 'torve'}`);
  } catch {
    console.error(`The Torvecode binary package ${packageName} is missing. Reinstall @torveai/cli without disabling optional dependencies.`);
    return 1;
  }

  // Interactive terminal programs must not run through spawnSync. On Windows,
  // the synchronous wrapper can keep the child from receiving terminal input,
  // resize and Ctrl+C events, leaving users on a blank screen while the native
  // process is alive. Keep npm's tiny launcher asynchronous and inherit the
  // terminal directly.
  return await new Promise((resolve) => {
    const child = spawn(executable, args, {
      stdio: 'inherit',
      windowsHide: false,
      windowsVerbatimArguments: false
    });
    child.once('error', (error) => {
      console.error(`Could not start Torvecode: ${error.message}`);
      resolve(1);
    });
    child.once('exit', (code, signal) => {
      if (signal && process.platform !== 'win32') {
        process.kill(process.pid, signal);
        return;
      }
      resolve(code == null ? 1 : code);
    });
  });
}

main()
  .then((status) => { process.exitCode = status; })
  .catch((error) => {
    console.error(`Update failed: ${error.message}`);
    process.exitCode = 1;
  });
