#!/usr/bin/env node

import { generateBinPath } from './node-platform'
const binPath = generateBinPath()

// eslint-disable-next-line @typescript-eslint/no-var-requires
require('child_process').execFileSync(binPath, process.argv.slice(2), { stdio: 'inherit' })
