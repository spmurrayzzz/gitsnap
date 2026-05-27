const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

const ext = {
  darwin: 'dylib',
  linux: 'so',
  win32: 'dll'
}[process.platform];

if (!ext) {
  throw new Error(`unsupported platform: ${process.platform}`);
}

fs.mkdirSync('bin', { recursive: true });

execFileSync('go', [
  'build',
  '-buildmode=c-shared',
  '-o',
  path.join('bin', `libgitsnap.${ext}`),
  './cmd/gitsnaplib'
], {
  stdio: 'inherit'
});
