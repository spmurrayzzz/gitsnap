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
```

Use another worktree with:

```sh
gitsnap --worktree /path/to/project save --alias checkpoint
```

## Storage

Snapshots are stored per worktree under:

- `$GITSNAP_HOME`, if set
- `$XDG_DATA_HOME/gitsnap`, if set
- `~/.local/share/gitsnap`, otherwise
