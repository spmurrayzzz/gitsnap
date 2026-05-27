const os = require('node:os');
const path = require('node:path');
const koffi = require('koffi');

const lib = koffi.load(libraryPath());
const api = {
  init: lib.func('void *GitsnapInit(const char *worktree)'),
  cleanup: lib.func('void *GitsnapCleanup(const char *worktree)'),
  save: lib.func('void *GitsnapSave(const char *worktree, const char *name)'),
  resolve: lib.func('void *GitsnapResolve(const char *worktree, const char *name)'),
  diff: lib.func('void *GitsnapDiff(const char *worktree, const char *name)'),
  files: lib.func('void *GitsnapFiles(const char *worktree, const char *name)'),
  restore: lib.func(
    'void *GitsnapRestore(const char *worktree, const char *name, const char *paths)'
  ),
  aliases: lib.func('void *GitsnapAliases(const char *worktree)'),
  free: lib.func('void GitsnapFree(void *ptr)')
};

function init(options = {}) {
  call(api.init, [worktree(options)]);
}

function cleanup(options = {}) {
  call(api.cleanup, [worktree(options)]);
}

function save(options = {}) {
  return call(api.save, [worktree(options), options.alias || '']);
}

function resolve(ref, options = {}) {
  return call(api.resolve, [worktree(options), ref]);
}

function diff(ref, options = {}) {
  return call(api.diff, [worktree(options), ref]);
}

function files(ref, options = {}) {
  return call(api.files, [worktree(options), ref]);
}

function restore(ref, options = {}) {
  const paths = JSON.stringify(options.paths || []);
  call(api.restore, [worktree(options), ref, paths]);
}

function aliases(options = {}) {
  return call(api.aliases, [worktree(options)]);
}

function call(fn, args) {
  const out = fn(...args);
  if (!out) {
    throw new Error('gitsnap returned a null response');
  }
  try {
    const text = koffi.decode(out, 'char', -1);
    const res = JSON.parse(text);
    if (!res.ok) {
      throw new Error(res.error || 'gitsnap failed');
    }
    return res.value;
  } finally {
    api.free(out);
  }
}

function worktree(options) {
  return options.worktree || '';
}

function libraryPath() {
  if (process.env.GITSNAP_LIB) {
    return process.env.GITSNAP_LIB;
  }
  const ext = {
    darwin: 'dylib',
    linux: 'so',
    win32: 'dll'
  }[process.platform];
  if (!ext) {
    throw new Error(`unsupported platform: ${process.platform}`);
  }
  return path.join(__dirname, '..', 'bin', `libgitsnap.${ext}`);
}

module.exports = {
  init,
  cleanup,
  save,
  resolve,
  diff,
  files,
  restore,
  aliases
};
