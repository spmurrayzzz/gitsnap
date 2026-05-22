package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spmurray/gitsnap/internal/alias"
	"github.com/spmurray/gitsnap/internal/backend/gogit"
	"github.com/spmurray/gitsnap/internal/store"
)

func TestMainErrorExit(t *testing.T) {
	oldArgs := os.Args
	oldExit := exit
	defer func() { os.Args = oldArgs; exit = oldExit }()
	os.Args = []string{"gitsnap", "--bad"}
	exit = func(code int) { panic(code) }
	defer func() {
		got := recover()
		if got != 1 {
			t.Fatalf("exit code = %#v", got)
		}
	}()
	main()
}

func TestRunForWorktreeError(t *testing.T) {
	old := forWorktree
	defer func() { forWorktree = old }()
	forWorktree = func(string) (store.WorktreeStore, error) {
		return store.WorktreeStore{}, errors.New("bad worktree")
	}
	if err := run(context.Background(), []string{"init"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunStoreEnsureError(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(home, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITSNAP_HOME", home)
	if err := run(context.Background(), []string{"init"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunResolveAndBackendErrors(t *testing.T) {
	t.Setenv("GITSNAP_HOME", filepath.Join(t.TempDir(), "home"))
	worktree := t.TempDir()
	for _, args := range [][]string{
		{"resolve", "missing"},
		{"diff", "missing"},
		{"files", "missing"},
		{"restore", "missing"},
	} {
		full := append([]string{"--worktree", worktree}, args...)
		if err := run(context.Background(), full); err == nil {
			t.Fatalf("expected error for %#v", full)
		}
	}
}

func TestRunCanceledBackendErrors(t *testing.T) {
	t.Setenv("GITSNAP_HOME", filepath.Join(t.TempDir(), "home"))
	worktree := t.TempDir()
	write(t, worktree, "a.txt", "hello\n")
	if err := run(context.Background(), []string{"--worktree", worktree, "init"}); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), []string{"--worktree", worktree, "save", "--alias", "first"}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, args := range [][]string{
		{"save"},
		{"diff", "first"},
		{"files", "first"},
		{"restore", "first"},
	} {
		full := append([]string{"--worktree", worktree}, args...)
		if err := run(ctx, full); err == nil {
			t.Fatalf("expected error for %#v", full)
		}
	}
}

func TestSaveAndListAliasErrors(t *testing.T) {
	ws, _ := commandDeps(t)
	write(t, ws.Worktree, "a.txt", "hello\n")
	badAliases := alias.Store{Path: t.TempDir()}
	if err := save(context.Background(), gogit.Backend{}, badAliases, ws, []string{"--alias", "bad"}); err == nil {
		t.Fatal("expected alias save error")
	}
	if err := listAliases(badAliases); err == nil {
		t.Fatal("expected alias list error")
	}
}

func TestSaveBackendError(t *testing.T) {
	ws, aliases := commandDeps(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := save(ctx, gogit.Backend{}, aliases, ws, nil); err == nil {
		t.Fatal("expected backend error")
	}
}

func TestRestoreResolveError(t *testing.T) {
	ws, aliases := commandDeps(t)
	if err := restore(context.Background(), gogit.Backend{}, aliases, ws, []string{"missing"}); err == nil {
		t.Fatal("expected resolve error")
	}
}
