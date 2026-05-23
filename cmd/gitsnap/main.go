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
	quiet := fs.Bool("quiet", false, "suppress status output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	args = fs.Args()
	if len(args) == 0 {
		usage()
		return nil
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
		if err := ws.Cleanup(); err != nil {
			return err
		}
		statusf(*quiet, "removed gitsnap storage\n")
		return nil
	}
	if args[0] == "init" {
		if len(args) != 1 {
			return fmt.Errorf("usage: gitsnap init")
		}
		if err := ws.Ensure(); err != nil {
			return err
		}
		err := backend.Init(ctx, ws.Worktree, ws.RepoDir())
		if err == nil {
			statusf(*quiet, "initialized gitsnap storage\n")
		}
		return err
	}
	if !needsInitialized(args[0]) {
		return fmt.Errorf("unknown command %q", args[0])
	}
	initialized, err := ws.Initialized()
	if err != nil {
		return err
	}
	if !initialized {
		return fmt.Errorf("worktree has not been initialized; run gitsnap init")
	}
	aliases := alias.Store{Path: ws.AliasPath()}

	if args[0] == "save" {
		return save(ctx, backend, aliases, ws, args[1:], *quiet)
	}
	if args[0] == "resolve" {
		if len(args) != 2 {
			return fmt.Errorf("usage: gitsnap resolve <alias>")
		}
		h, err := ref.Resolve(aliases, args[1])
		if err != nil {
			return err
		}
		fmt.Println(h)
		return nil
	}
	if args[0] == "aliases" {
		return listAliases(aliases, *quiet)
	}
	if args[0] == "diff" {
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
		if len(out) == 0 {
			statusf(*quiet, "no changes\n")
			return nil
		}
		fmt.Print(string(out))
		return nil
	}
	if args[0] == "files" {
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
		if len(files) == 0 {
			statusf(*quiet, "no changed files\n")
			return nil
		}
		for _, file := range files {
			fmt.Println(file)
		}
		return nil
	}
	return restore(ctx, backend, aliases, ws, args[1:], *quiet)
}

func save(
	ctx context.Context,
	backend gogit.Backend,
	aliases alias.Store,
	ws store.WorktreeStore,
	args []string,
	quiet bool,
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
	statusf(quiet, "saved snapshot %s\n", h)
	if *name != "" {
		statusf(quiet, "updated alias %s -> %s\n", *name, h)
	}
	return nil
}

func restore(
	ctx context.Context,
	backend gogit.Backend,
	aliases alias.Store,
	ws store.WorktreeStore,
	args []string,
	quiet bool,
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
	if err := backend.Restore(ctx, ws.Worktree, ws.RepoDir(), h, paths); err != nil {
		return err
	}
	statusf(quiet, "restored snapshot %s\n", h)
	return nil
}

func statusf(quiet bool, format string, args ...any) {
	if !quiet {
		fmt.Printf(format, args...)
	}
}

func listAliases(aliases alias.Store, quiet bool) error {
	names, records, err := aliases.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		statusf(quiet, "no aliases\n")
		return nil
	}
	for _, name := range names {
		rec := records[name]
		fmt.Printf("%s %s %s\n", name, rec.Hash, rec.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
	return nil
}

func needsInitialized(cmd string) bool {
	switch cmd {
	case "save", "diff", "files", "restore", "resolve", "aliases":
		return true
	default:
		return false
	}
}

func usage() {
	fmt.Println(strings.TrimSpace(`gitsnap [--worktree PATH] [--quiet] <command>

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
