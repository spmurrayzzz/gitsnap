# gitsnap

Minimal, git-driven local project snapshots.

`gitsnap` saves the current file tree outside the repo, then lets you diff,
list, resolve, and restore those snapshots later. It is useful for quick
checkpoints that should not become commits, stashes, or branches.

## Build

Building the CLI requires Go 1.25.1 or newer.

Building the Node.js FFI shared library also requires CGO and a native C
compiler toolchain:

- macOS: Xcode Command Line Tools
- Linux: GCC or Clang with libc development headers
- Windows: a CGO-compatible C compiler such as MinGW-w64

Installing the Node.js package from GitHub also requires Node.js, npm, Go, and
that same native C compiler toolchain because the shared library is built during
install.

```sh
make build
```

This writes `bin/gitsnap`.

Build the shared library for Node.js FFI bindings with:

```sh
make build-lib
```

This writes `bin/libgitsnap.dylib`, `bin/libgitsnap.so`, or
`bin/libgitsnap.dll`, depending on the platform.

## Usage

```sh
gitsnap init
gitsnap save --alias before-refactor
gitsnap diff before-refactor
gitsnap files before-refactor
gitsnap restore before-refactor -- path/to/file
gitsnap restore before-refactor
gitsnap aliases
gitsnap resolve before-refactor
gitsnap cleanup
```

Use another worktree with:

```sh
gitsnap --worktree /path/to/project save --alias checkpoint
```

## Node.js

This repository includes a Node.js FFI API and a pi skill under
`skills/gitsnap-node/` to help AI agents use the library correctly.

Install from GitHub with:

```sh
npm install github:spmurrayzzz/gitsnap
```

The install builds the native shared library locally, so the build toolchains
listed above must be available.

```js
const gitsnap = require("gitsnap");

gitsnap.init({ worktree: "/path/to/project" });

const hash = gitsnap.save({
  worktree: "/path/to/project",
  alias: "checkpoint"
});

console.log(hash);
console.log(gitsnap.aliases({ worktree: "/path/to/project" }));
```

## How it works

`gitsnap` uses `go-git` to maintain a separate Git object database for each
worktree. A snapshot is a Git tree object, not a commit: `save` indexes the
current non-ignored worktree files, writes blobs and trees into gitsnap storage,
and prints the root tree hash.

`diff` and `files` save the current tree, then compare it with the requested
snapshot. `restore` reads files from a saved tree back into the worktree; a full
restore also removes non-ignored files that are absent from the saved tree.
Aliases are just names stored in `aliases.json` that point at tree hashes.

## Storage

Snapshots are stored per worktree under:

- `$GITSNAP_HOME`, if set
- `$XDG_DATA_HOME/gitsnap`, if set
- `~/.local/share/gitsnap`, otherwise

Use `gitsnap cleanup` to remove all gitsnap data for the current worktree.
