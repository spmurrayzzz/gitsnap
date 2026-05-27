# gitsnap node api

Node.js FFI bindings for `gitsnap`.

## install

```sh
npm install github:spmurrayzzz/gitsnap
```

GitHub installs build the native shared library locally. Go and a CGO-capable C
compiler must be available.

## usage

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

All methods are synchronous.

## api

### init(options)

Initializes gitsnap storage for a worktree.

```js
gitsnap.init({ worktree: "/path/to/project" });
```

### save(options)

Saves the current worktree and returns the snapshot hash.

```js
const hash = gitsnap.save({
  worktree: "/path/to/project",
  alias: "before-refactor"
});
```

`alias` is optional.

### resolve(ref, options)

Resolves an alias or hash and returns the snapshot hash.

```js
const hash = gitsnap.resolve("before-refactor", { worktree });
```

### diff(ref, options)

Returns a git-style diff between the current worktree and a snapshot.

```js
const patch = gitsnap.diff("before-refactor", { worktree });
```

### files(ref, options)

Returns changed file paths between the current worktree and a snapshot.

```js
const files = gitsnap.files("before-refactor", { worktree });
```

### restore(ref, options)

Restores a full snapshot, or only selected paths.

```js
gitsnap.restore("before-refactor", { worktree });

gitsnap.restore("before-refactor", {
  worktree,
  paths: ["src/index.js"]
});
```

### aliases(options)

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

### cleanup(options)

Removes all gitsnap storage for a worktree.

```js
gitsnap.cleanup({ worktree });
```

## options

All methods that take `options` accept:

```ts
{
  worktree?: string;
}
```

If `worktree` is omitted, gitsnap uses the current working directory.
