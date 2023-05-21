import child_process = require('child_process')
// import fs = require('fs')
import path = require('path')
// import tty = require('tty')

import { YALIVE_BINARY_PATH, generateBinPath } from './npm/node-platform'
import * as common from './shared/common'
import * as types from './shared/types'

// This file is used for both the "yalive-server" package and the "yalive-server-wasm"
// package. "WASM" will be true for "yalive-server-wasm" and false for "yalive-server".
declare const WASM: boolean

const defaultWD = process.cwd()
const isRunning = false

const yaliveCommandAndArgs = (): [string, string[]] => {
  // Try to have a nice error message when people accidentally bundle yalive-server
  // without providing an explicit path to the binary, or when using WebAssembly.
  if (
    (!YALIVE_BINARY_PATH || WASM) &&
    (path.basename(__filename) !== 'main.js' || path.basename(__dirname) !== 'lib')
  ) {
    throw new Error(
      /* eslint-disable quotes */
      `The yalive-server JavaScript API cannot be bundled. Please mark the "yalive-server" package as external so it's not included in the bundle.\n\nMore information: The file containing the code for yalive-server's JavaScript API (${__filename}) does not appear to be inside the yalive-server package on the file system, which usually means that the yalive-server package was bundled into another file. This is problematic because the API needs to run a binary executable inside the yalive-server package which is located using a relative path from the API code to the executable. If the yalive-server package is bundled, the relative path will be incorrect and the executable won't be found.`,
      /* eslint-enable quotes */
    )
  }

  if (WASM) {
    return ['node', [path.join(__dirname, '..', 'bin', 'yalive-server')]]
  }

  return [generateBinPath(), []]
}

export const dev: typeof types.dev = (options: types.DevOptions): Promise<void> => {
  if (isRunning) {
    return Promise.resolve()
  }

  const [command, args] = yaliveCommandAndArgs()

  // TODO: validate options
  const configStr = JSON.stringify(options)

  const child = child_process.spawn(command, args.concat('dev', '-c', configStr), {
    windowsHide: true,
    stdio: ['pipe', 'pipe', 'inherit'],
    cwd: defaultWD,
  })

  const stdin: typeof child.stdin & { unref?(): void } = child.stdin
  const stdout: typeof child.stdout & { unref?(): void } = child.stdout

  stdout.on('data', (chunk) => {
    console.debug(chunk)
  })
  stdout.on('end', () => {
    console.log('server closed')
  })

  let refCount = 0
  child.unref()
  if (stdin.unref) {
    stdin.unref()
  }
  if (stdout.unref) {
    stdout.unref()
  }

  const _refs: common.Refs = {
    ref() {
      if (++refCount === 1) child.ref()
    },
    unref() {
      if (--refCount === 0) child.unref()
    },
  }

  return Promise.resolve()
}
