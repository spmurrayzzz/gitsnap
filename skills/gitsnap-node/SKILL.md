---
name: gitsnap-node
description: Helps agents use the gitsnap Node.js FFI API for local project snapshots. Use when writing JavaScript or TypeScript code that imports gitsnap, calls init/save/diff/files/restore/aliases/cleanup, or installs gitsnap from GitHub with npm.
---

# Gitsnap Node.js API

## Quick start

Install from GitHub:

```sh
npm install github:spmurrayzzz/gitsnap
```

GitHub installs build the native shared library locally. The target machine must
have Go and a CGO-capable C compiler.

Use the same `worktree` for every operation:

```js
import gitsnap from "gitsnap";

const worktree = "/path/to/project";

gitsnap.init({ worktree });

const hash = gitsnap.save({
  worktree,
  alias: "checkpoint"
});

console.log(hash);
console.log(gitsnap.aliases({ worktree }));
```

All methods are synchronous. If `worktree` is omitted, gitsnap uses the current
working directory.

## API

### `init(options)`

Initializes gitsnap storage for a worktree.

```js
gitsnap.init({ worktree });
```

### `save(options)`

Saves the current worktree and returns the snapshot hash. `alias` is optional.

```js
const hash = gitsnap.save({ worktree, alias: "before-refactor" });
```

### `resolve(ref, options)`

Resolves an alias or hash and returns the snapshot hash.

```js
const hash = gitsnap.resolve("before-refactor", { worktree });
```

### `diff(ref, options)`

Returns a git-style diff between the current worktree and a snapshot.

```js
const patch = gitsnap.diff("before-refactor", { worktree });
```

### `files(ref, options)`

Returns changed file paths between the current worktree and a snapshot.

```js
const files = gitsnap.files("before-refactor", { worktree });
```

### `restore(ref, options)`

Restores a full snapshot, or only selected paths.

```js
gitsnap.restore("before-refactor", { worktree });

gitsnap.restore("before-refactor", {
  worktree,
  paths: ["src/index.js"]
});
```

### `aliases(options)`

Returns aliases sorted by name.

```js
const aliases = gitsnap.aliases({ worktree });
```

Each alias has this shape:

```ts
{
  name: string;
  hash: string;
  updated_at: string;
}
```

### `cleanup(options)`

Removes all gitsnap storage for a worktree.

```js
gitsnap.cleanup({ worktree });
```

## Troubleshooting

If `save`, `diff`, `files`, `restore`, `aliases`, or `resolve` fails with
`worktree has not been initialized; run gitsnap init`, call `init` with the same
`worktree` path used by the later operation.

Bad:

```js
gitsnap.init({ worktree: "/path/to/repo" });
gitsnap.save({ worktree: "/Users/me/project" });
```

Good:

```js
const worktree = "/Users/me/project";
gitsnap.init({ worktree });
gitsnap.save({ worktree });
```
