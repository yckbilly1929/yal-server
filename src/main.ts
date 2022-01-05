import child_process = require('child_process')
// import fs = require('fs')
import path = require('path')
// import tty = require('tty')

import { generateBinPath, YAL_BINARY_PATH } from "./npm/node-platform"
import * as common from "./shared/common"
import * as types from "./shared/types"

// This file is used for both the "yal-server" package and the "yal-server-wasm"
// package. "WASM" will be true for "yal-server-wasm" and false for "yal-server".
declare const WASM: boolean;

let defaultWD = process.cwd()
let isRunning = false

let yalCommandAndArgs = (): [string, string[]] => {
  // Try to have a nice error message when people accidentally bundle yal-server
  // without providing an explicit path to the binary, or when using WebAssembly.
  if ((!YAL_BINARY_PATH || WASM) && (path.basename(__filename) !== 'main.js' || path.basename(__dirname) !== 'lib')) {
    throw new Error(
      `The yal-server JavaScript API cannot be bundled. Please mark the "yal-server" ` +
      `package as external so it's not included in the bundle.\n` +
      `\n` +
      `More information: The file containing the code for yal-server's JavaScript ` +
      `API (${__filename}) does not appear to be inside the yal-server package on ` +
      `the file system, which usually means that the yal-server package was bundled ` +
      `into another file. This is problematic because the API needs to run a ` +
      `binary executable inside the yal-server package which is located using a ` +
      `relative path from the API code to the executable. If the yal-server package ` +
      `is bundled, the relative path will be incorrect and the executable won't ` +
      `be found.`);
  }

  if (WASM) {
    return ['node', [path.join(__dirname, '..', 'bin', 'yal-server')]];
  }

  return [generateBinPath(), []];
};

export let dev: typeof types.dev = (options: types.DevOptions): Promise<void> => {
  if (isRunning) {
    return Promise.resolve()
  }

  let [command, args] = yalCommandAndArgs()

  // TODO: validate options
  const configStr = JSON.stringify(options)

  let child = child_process.spawn(command, args.concat('dev', '-c', configStr), {
    windowsHide: true,
    stdio: ['pipe', 'pipe', 'inherit'],
    cwd: defaultWD,
  });
  
  const stdin: typeof child.stdin & { unref?(): void } = child.stdin;
  const stdout: typeof child.stdout & { unref?(): void } = child.stdout;

  stdout.on('data', (chunk) => {
    console.debug(chunk)
  });
  stdout.on('end', () => {
    console.log('server closed')
  })

  let refCount = 0;
  child.unref();
  if (stdin.unref) {
    stdin.unref();
  }
  if (stdout.unref) {
    stdout.unref();
  }

  const refs: common.Refs = {
    ref() { if (++refCount === 1) child.ref(); },
    unref() { if (--refCount === 0) child.unref(); },
  }

  return Promise.resolve()
}
