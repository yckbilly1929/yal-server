import fs = require('fs');
import os = require('os');
import path = require('path');

// This feature was added to give external code a way to modify the binary
// path without modifying the code itself. Do not remove this because
// external code relies on this.
export var YALIVE_BINARY_PATH: string | undefined = process.env.YALIVE_BINARY_PATH || YALIVE_BINARY_PATH;

export const knownWindowsPackages: Record<string, string> = {
  // 'win32 arm64 LE': 'yalive-server-windows-arm64',
  // 'win32 ia32 LE': 'yalive-server-windows-32',
  // 'win32 x64 LE': 'yalive-server-windows-64',
};

export const knownUnixlikePackages: Record<string, string> = {
  'android arm64 LE': 'yalive-server-android-arm64',
  'darwin arm64 LE': 'yalive-server-darwin-arm64',
  'darwin x64 LE': 'yalive-server-darwin-64',
  'freebsd arm64 LE': 'yalive-server-freebsd-arm64',
  'freebsd x64 LE': 'yalive-server-freebsd-64',
  'linux arm LE': 'yalive-server-linux-arm',
  'linux arm64 LE': 'yalive-server-linux-arm64',
  'linux ia32 LE': 'yalive-server-linux-32',
  'linux mips64el LE': 'yalive-server-linux-mips64le',
  'linux ppc64 LE': 'yalive-server-linux-ppc64le',
  'linux s390x BE': 'yalive-server-linux-s390x',
  'linux x64 LE': 'yalive-server-linux-64',
  'netbsd x64 LE': 'yalive-server-netbsd-64',
  'openbsd x64 LE': 'yalive-server-openbsd-64',
  'sunos x64 LE': 'yalive-server-sunos-64',
};

export function pkgAndSubpathForCurrentPlatform(): { pkg: string, subpath: string } {
  let pkg: string;
  let subpath: string;
  let platformKey = `${process.platform} ${os.arch()} ${os.endianness()}`;

  if (platformKey in knownWindowsPackages) {
    pkg = knownWindowsPackages[platformKey];
    subpath = 'yalive-server.exe';
  }

  else if (platformKey in knownUnixlikePackages) {
    // pkg = knownUnixlikePackages[platformKey];
    pkg = 'yalive-server'
    subpath = 'bin/yalive-server';
  }

  else {
    throw new Error(`Unsupported platform: ${platformKey}`);
  }

  return { pkg, subpath };
}

export function downloadedBinPath(pkg: string, subpath: string): string {
  const libDir = path.dirname(require.resolve('yalive-server'));
  return path.join(libDir, `downloaded-${pkg}-${path.basename(subpath)}`);
}

export function generateBinPath(): string {
  // This feature was added to give external code a way to modify the binary
  // path without modifying the code itself. Do not remove this because
  // external code relies on this (in addition to yalive-server's own test suite).
  if (YALIVE_BINARY_PATH) {
    return YALIVE_BINARY_PATH;
  }

  const { pkg, subpath } = pkgAndSubpathForCurrentPlatform();
  let binPath: string;

  try {
    // First check for the binary package from our "optionalDependencies". This
    // package should have been installed alongside this package at install time.
    binPath = require.resolve(`${pkg}/${subpath}`);
  } catch (e) {
    // If that didn't work, then someone probably installed yalive-server with the
    // "--no-optional" flag. Our install script attempts to compensate for this
    // by manually downloading the package instead. Check for that next.
    binPath = downloadedBinPath(pkg, subpath);
    if (!fs.existsSync(binPath)) {
      // If that didn't work too, then we're out of options. This can happen
      // when someone installs yalive-server with both the "--no-optional" and the
      // "--ignore-scripts" flags. The fix for this is to just not do that.
      //
      // In that case we try to have a nice error message if we think we know
      // what's happening. Otherwise we just rethrow the original error message.
      try {
        require.resolve(pkg);
      } catch {
        throw new Error(`The package "${pkg}" could not be found, and is needed by yalive-server.

If you are installing yalive-server with npm, make sure that you don't specify the
"--no-optional" flag. The "optionalDependencies" package.json feature is used
by yalive-server to install the correct binary executable for your current platform.`);
      }
      throw e;
    }
  }

  // The yalive-server binary executable can't be used in Yarn 2 in PnP mode because
  // it's inside a virtual file system and the OS needs it in the real file
  // system. So we need to copy the file out of the virtual file system into
  // the real file system.
  let isYarnPnP = false;
  try {
    require('pnpapi');
    isYarnPnP = true;
  } catch (e) {
  }
  if (isYarnPnP) {
    const libDir = path.dirname(require.resolve('yalive-server'));
    const binTargetPath = path.join(libDir, `pnpapi-${pkg}-${path.basename(subpath)}`);
    if (!fs.existsSync(binTargetPath)) {
      fs.copyFileSync(binPath, binTargetPath);
      fs.chmodSync(binTargetPath, 0o755);
    }
    return binTargetPath;
  }

  return binPath;
}
