# gitsnap

Minimal, git-driven local project snapshots.

`gitsnap` saves the current file tree outside the repo, then lets you diff,
list, resolve, and restore those snapshots later. It is useful for quick
checkpoints that should not become commits, stashes, or branches.

## Build

```sh
make build
```

This writes `bin/gitsnap`.

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
