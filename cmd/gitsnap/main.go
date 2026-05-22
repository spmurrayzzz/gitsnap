package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spmurray/gitsnap/internal/alias"
	"github.com/spmurray/gitsnap/internal/backend/gogit"
	"github.com/spmurray/gitsnap/internal/ref"
	"github.com/spmurray/gitsnap/internal/store"
)

var exit = os.Exit
var forWorktree = store.ForWorktree

func main() {
	code := 0
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gitsnap:", err)
		code = 1
	}
	exit(code)
}

func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("gitsnap", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	worktree := fs.String("worktree", ".", "worktree path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	args = fs.Args()
	if len(args) == 0 {
		usage()
		return nil
	}
	if !knownCommand(args[0]) {
		return fmt.Errorf("unknown command %q", args[0])
	}

	ws, err := forWorktree(*worktree)
	if err != nil {
		return err
	}
	backend := gogit.Backend{}
	if args[0] == "cleanup" {
		if len(args) != 1 {
			return fmt.Errorf("usage: gitsnap cleanup")
		}
		return ws.Cleanup()
	}
	if args[0] == "init" {
		if len(args) != 1 {
			return fmt.Errorf("usage: gitsnap init")
		}
		if err := ws.Ensure(); err != nil {
			return err
		}
		return backend.Init(ctx, ws.Worktree, ws.RepoDir())
	}
	initialized, err := ws.Initialized()
	if err != nil {
		return err
	}
	if !initialized {
		return fmt.Errorf("worktree has not been initialized; run gitsnap init")
	}
	aliases := alias.Store{Path: ws.AliasPath()}

	switch args[0] {
	case "save":
		return save(ctx, backend, aliases, ws, args[1:])
	case "resolve":
		if len(args) != 2 {
			return fmt.Errorf("usage: gitsnap resolve <alias>")
		}
		h, err := ref.Resolve(aliases, args[1])
		if err != nil {
			return err
		}
		fmt.Println(h)
		return nil
	case "aliases":
		return listAliases(aliases)
	case "diff":
		if len(args) != 2 {
			return fmt.Errorf("usage: gitsnap diff <snapshot-or-alias>")
		}
		h, err := ref.Resolve(aliases, args[1])
		if err != nil {
			return err
		}
		out, err := backend.Diff(ctx, ws.Worktree, ws.RepoDir(), h)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	case "files":
		if len(args) != 2 {
			return fmt.Errorf("usage: gitsnap files <snapshot-or-alias>")
		}
		h, err := ref.Resolve(aliases, args[1])
		if err != nil {
			return err
		}
		files, err := backend.ChangedFiles(ctx, ws.Worktree, ws.RepoDir(), h)
		if err != nil {
			return err
		}
		for _, file := range files {
			fmt.Println(file)
		}
		return nil
	case "restore":
		return restore(ctx, backend, aliases, ws, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func save(
	ctx context.Context,
	backend gogit.Backend,
	aliases alias.Store,
	ws store.WorktreeStore,
	args []string,
) error {
	fs := flag.NewFlagSet("save", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	name := fs.String("alias", "", "alias name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: gitsnap save [--alias NAME]")
	}
	h, err := backend.Save(ctx, ws.Worktree, ws.RepoDir())
	if err != nil {
		return err
	}
	if *name != "" {
		if err := aliases.Set(*name, h); err != nil {
			return err
		}
	}
	fmt.Println(h)
	return nil
}

func restore(
	ctx context.Context,
	backend gogit.Backend,
	aliases alias.Store,
	ws store.WorktreeStore,
	args []string,
) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gitsnap restore <snapshot-or-alias> [-- PATH...]")
	}
	h, err := ref.Resolve(aliases, args[0])
	if err != nil {
		return err
	}
	paths := args[1:]
	if len(paths) > 0 && paths[0] == "--" {
		paths = paths[1:]
	}
	return backend.Restore(ctx, ws.Worktree, ws.RepoDir(), h, paths)
}

func listAliases(aliases alias.Store) error {
	names, records, err := aliases.List()
	if err != nil {
		return err
	}
	for _, name := range names {
		rec := records[name]
		fmt.Printf("%s %s %s\n", name, rec.Hash, rec.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
	return nil
}

func knownCommand(cmd string) bool {
	switch cmd {
	case "init", "cleanup", "save", "diff", "files", "restore",
		"resolve", "aliases":
		return true
	default:
		return false
	}
}

func usage() {
	fmt.Println(strings.TrimSpace(`gitsnap [--worktree PATH] <command>

commands:
  init
  cleanup
  save [--alias NAME]
  diff <snapshot-or-alias>
  files <snapshot-or-alias>
  restore <snapshot-or-alias> [-- PATH...]
  resolve <alias>
  aliases`))
}
