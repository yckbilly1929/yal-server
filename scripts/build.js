/* eslint-disable @typescript-eslint/no-var-requires */
const childProcess = require('child_process')
const path = require('path')
const fs = require('fs')

const repoDir = path.dirname(__dirname)
const npmDir = path.join(repoDir, 'npm', 'yalive-server')
const version = fs.readFileSync(path.join(repoDir, 'version.txt'), 'utf8').trim()
const nodeTarget = 'node16' // See: https://nodejs.org/en/about/releases/

const buildNeutralLib = () => {
  const libDir = path.join(npmDir, 'lib')
  const binDir = path.join(npmDir, 'bin')
  fs.mkdirSync(libDir, { recursive: true })
  fs.mkdirSync(binDir, { recursive: true })

  // Generate "npm/yalive-server/install.js"
  childProcess.execSync(
    [
      'pnpm exec esbuild',
      path.join(repoDir, 'src', 'npm', 'node-install.ts'),
      '--outfile=' + path.join(npmDir, 'install.js'),
      '--bundle',
      '--define:YALIVE_VERSION=' + `\\"${version}\\"`,
      '--external:yalive-server',
      '--platform=node',
      '--log-level=warning',
    ].join(' ')
  )

  // Generate "npm/yalive-server/lib/main.js"
  childProcess.execSync(
    [
      'pnpm exec esbuild',
      path.join(repoDir, 'src', 'main.ts'),
      '--outfile=' + path.join(libDir, 'main.js'),
      '--bundle',
      '--define:WASM=false',
      '--define:YALIVE_VERSION=' + `\\"${version}\\"`,
      '--external:yalive-server',
      '--platform=node',
      '--log-level=warning',
    ].join(' ')
  )

  // Generate "npm/yalive-server/lib/main.d.ts"
  const types_ts = fs.readFileSync(path.join(repoDir, 'src', 'shared', 'types.ts'), 'utf8')
  fs.writeFileSync(path.join(libDir, 'main.d.ts'), types_ts)

  // Get supported platforms
  const platforms = { exports: {} }
  new Function(
    'module',
    'exports',
    'require',
    childProcess.execSync(
      [
        'pnpm exec esbuild',
        path.join(repoDir, 'src', 'npm', 'node-platform.ts'),
        '--bundle',
        '--target=' + nodeTarget,
        '--external:yalive-server',
        '--platform=node',
        '--log-level=warning',
      ].join(' ')
    )
  )(platforms, platforms.exports, require)
  const optionalDependencies = Object.fromEntries(
    Object.values({
      ...platforms.exports.knownWindowsPackages,
      ...platforms.exports.knownUnixlikePackages,
      ...platforms.exports.knownWebAssemblyFallbackPackages,
    })
      .sort()
      .map((x) => [x, version])
  )

  // Update "npm/yalive-server/package.json"
  const pjPath = path.join(npmDir, 'package.json')
  const package_json = JSON.parse(fs.readFileSync(pjPath, 'utf8'))
  package_json.optionalDependencies = optionalDependencies
  fs.writeFileSync(pjPath, JSON.stringify(package_json, null, 2) + '\n')
}

// exports.buildBinary = () => {
//   childProcess.execFileSync('go', ['build', '-ldflags=-s -w', './cmd/root'], { cwd: repoDir, stdio: 'ignore' })
//   return path.join(repoDir, process.platform === 'win32' ? 'esbuild.exe' : 'esbuild')
// }

exports.removeRecursiveSync = (p) => {
  try {
    fs.rmSync(p, { recursive: true })
  } catch (e) {
    // Removing stuff on Windows is flaky and unreliable. Don't fail tests
    // on CI if Windows is just being a pain. Common causes of flakes include
    // random EPERM and ENOTEMPTY errors.
    //
    // The general "solution" to this is to try asking Windows to redo the
    // failing operation repeatedly until eventually giving up after a
    // timeout. But that doesn't guarantee that flakes will be fixed so we
    // just give up instead. People that want reasonable file system
    // behavior on Windows should use WSL instead.
  }
}

const updateVersionPackageJSON = (pathToPackageJSON) => {
  const v = fs.readFileSync(path.join(path.dirname(__dirname), 'version.txt'), 'utf8').trim()
  const json = JSON.parse(fs.readFileSync(pathToPackageJSON, 'utf8'))
  if (json.version !== v) {
    json.version = v
    fs.writeFileSync(pathToPackageJSON, JSON.stringify(json, null, 2) + '\n')
  }
}

// exports.installForTests = () => {
//   // Build the "esbuild" binary and library
//   const buildPath = exports.buildBinary()
//   buildNeutralLib(buildPath)

//   // Install the "esbuild" package to a temporary directory. On Windows, it's
//   // sometimes randomly impossible to delete this installation directory. My
//   // best guess is that this is because the esbuild process is kept alive until
//   // the process exits for "buildSync" and "transformSync", and that sometimes
//   // prevents Windows from deleting the directory it's in. The call in tests to
//   // "rimraf.sync()" appears to hang when this happens. Other operating systems
//   // don't have a problem with this. This has only been a problem on the Windows
//   // VM in GitHub CI. I cannot reproduce this issue myself.
//   const installDir = path.join(os.tmpdir(), 'esbuild-' + Math.random().toString(36).slice(2))
//   const env = { ...process.env, ESBUILD_BINARY_PATH: buildPath }
//   fs.mkdirSync(installDir)
//   fs.writeFileSync(path.join(installDir, 'package.json'), '{}')
//   childProcess.execSync(`npm pack --silent "${npmDir}"`, { cwd: installDir, stdio: 'inherit' })
//   childProcess.execSync(`npm install --silent --no-audit --progress=false esbuild-${version}.tgz`, {
//     cwd: installDir,
//     env,
//     stdio: 'inherit',
//   })

//   // Evaluate the code
//   const ESBUILD_PACKAGE_PATH = path.join(installDir, 'node_modules', 'esbuild')
//   const mod = require(ESBUILD_PACKAGE_PATH)
//   Object.defineProperty(mod, 'ESBUILD_PACKAGE_PATH', { value: ESBUILD_PACKAGE_PATH })
//   return mod
// }

const updateVersionGo = () => {
  const version_txt = fs.readFileSync(path.join(repoDir, 'version.txt'), 'utf8').trim()
  const version_go = `package cmd\n\nconst yaliveVersion = "${version_txt}"\n`
  const version_go_path = path.join(repoDir, 'cmd', 'version.go')

  // Update this atomically to avoid issues with this being overwritten during use
  const temp_path = version_go_path + Math.random().toString(36).slice(1)
  fs.writeFileSync(temp_path, version_go)
  fs.renameSync(temp_path, version_go_path)
}

// This is helpful for ES6 modules which don't have access to __dirname
exports.dirname = __dirname

// The main Makefile invokes this script before publishing
if (require.main === module) {
  // if (process.argv.indexOf('--wasm') >= 0) exports.buildWasmLib(process.argv[2])
  // else if (process.argv.indexOf('--deno') >= 0) buildDenoLib(process.argv[2])
  if (process.argv.indexOf('--version') >= 0) updateVersionPackageJSON(process.argv[2])
  else if (process.argv.indexOf('--neutral') >= 0) buildNeutralLib(process.argv[2])
  else if (process.argv.indexOf('--update-version-go') >= 0) updateVersionGo()
  else throw new Error('Expected a flag')
}
