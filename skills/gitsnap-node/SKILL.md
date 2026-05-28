---
name: gitsnap-node
description: Helps agents use the gitsnap Node.js FFI API for local project snapshots. Use when writing JavaScript or TypeScript code that imports gitsnap, calls init/save/diff/files/restore/aliases/cleanup, awaits gitsnap methods, uses gitsnap.sync, or installs gitsnap from GitHub with npm.
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

await gitsnap.init({ worktree });

const hash = await gitsnap.save({
  worktree,
  alias: "checkpoint"
});

console.log(hash);
console.log(await gitsnap.aliases({ worktree }));
```

The default API is async and returns promises. If `worktree` is omitted, gitsnap
uses the current working directory.

## API

### `init(options)`

Initializes gitsnap storage for a worktree.

```js
await gitsnap.init({ worktree });
```

### `save(options)`

Saves the current worktree and returns the snapshot hash. `alias` is optional.

```js
const hash = await gitsnap.save({ worktree, alias: "before-refactor" });
```

### `resolve(ref, options)`

Resolves an alias or hash and returns the snapshot hash.

```js
const hash = await gitsnap.resolve("before-refactor", { worktree });
```

### `diff(ref, options)`

Returns a git-style diff between the current worktree and a snapshot.

```js
const patch = await gitsnap.diff("before-refactor", { worktree });
```

### `files(ref, options)`

Returns changed file paths between the current worktree and a snapshot.

```js
const files = await gitsnap.files("before-refactor", { worktree });
```

### `restore(ref, options)`

Restores a full snapshot, or only selected paths.

```js
await gitsnap.restore("before-refactor", { worktree });

await gitsnap.restore("before-refactor", {
  worktree,
  paths: ["src/index.js"]
});
```

### `aliases(options)`

Returns aliases sorted by name.

```js
const aliases = await gitsnap.aliases({ worktree });
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
await gitsnap.cleanup({ worktree });
```

## Sync API

Blocking variants are available under `gitsnap.sync`.

```js
const hash = gitsnap.sync.save({ worktree, alias: "checkpoint" });
```

## Troubleshooting

If `save`, `diff`, `files`, `restore`, `aliases`, or `resolve` fails with
`worktree has not been initialized; run gitsnap init`, call `init` with the same
`worktree` path used by the later operation.

Bad:

```js
await gitsnap.init({ worktree: "/path/to/repo" });
await gitsnap.save({ worktree: "/Users/me/project" });
```

Good:

```js
const worktree = "/Users/me/project";
await gitsnap.init({ worktree });
await gitsnap.save({ worktree });
```
