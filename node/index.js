const fs = require('node:fs');
const path = require('node:path');
const koffi = require('koffi');

const lib = koffi.load(libraryPath());
const queues = new Map();
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

async function init(options = {}) {
  await callAsync(api.init, [worktree(options)]);
}

async function cleanup(options = {}) {
  await callAsync(api.cleanup, [worktree(options)]);
}

async function save(options = {}) {
  return callAsync(api.save, [worktree(options), options.alias || '']);
}

async function resolve(ref, options = {}) {
  return callAsync(api.resolve, [worktree(options), ref]);
}

async function diff(ref, options = {}) {
  return callAsync(api.diff, [worktree(options), ref]);
}

async function files(ref, options = {}) {
  return callAsync(api.files, [worktree(options), ref]);
}

async function restore(ref, options = {}) {
  const paths = JSON.stringify(options.paths || []);
  await callAsync(api.restore, [worktree(options), ref, paths]);
}

async function aliases(options = {}) {
  return callAsync(api.aliases, [worktree(options)]);
}

function initSync(options = {}) {
  call(api.init, [worktree(options)]);
}

function cleanupSync(options = {}) {
  call(api.cleanup, [worktree(options)]);
}

function saveSync(options = {}) {
  return call(api.save, [worktree(options), options.alias || '']);
}

function resolveSync(ref, options = {}) {
  return call(api.resolve, [worktree(options), ref]);
}

function diffSync(ref, options = {}) {
  return call(api.diff, [worktree(options), ref]);
}

function filesSync(ref, options = {}) {
  return call(api.files, [worktree(options), ref]);
}

function restoreSync(ref, options = {}) {
  const paths = JSON.stringify(options.paths || []);
  call(api.restore, [worktree(options), ref, paths]);
}

function aliasesSync(options = {}) {
  return call(api.aliases, [worktree(options)]);
}

function callAsync(fn, args) {
  return enqueue(queueKey(args[0]), () => new Promise((resolve, reject) => {
    fn.async(...args, (err, out) => {
      if (err) {
        reject(err);
        return;
      }
      try {
        resolve(read(out));
      } catch (readErr) {
        reject(readErr);
      }
    });
  }));
}

function enqueue(key, run) {
  const previous = queues.get(key) || Promise.resolve();
  const next = previous.catch(() => {}).then(run);
  queues.set(key, next);
  next.finally(() => {
    if (queues.get(key) === next) {
      queues.delete(key);
    }
  }).catch(() => {});
  return next;
}

function call(fn, args) {
  const key = queueKey(args[0]);
  if (queues.has(key)) {
    throw new Error('gitsnap async operation already pending for worktree');
  }
  return read(fn(...args));
}

function queueKey(value) {
  const target = path.resolve(value || '.');
  try {
    return fs.realpathSync.native(target);
  } catch {
    return target;
  }
}

function read(out) {
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
  aliases,
  sync: {
    init: initSync,
    cleanup: cleanupSync,
    save: saveSync,
    resolve: resolveSync,
    diff: diffSync,
    files: filesSync,
    restore: restoreSync,
    aliases: aliasesSync
  }
};
