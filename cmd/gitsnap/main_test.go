package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spmurray/gitsnap/internal/alias"
	"github.com/spmurray/gitsnap/internal/backend/gogit"
	"github.com/spmurray/gitsnap/internal/store"
)

func TestRunNoArgsShowsUsage(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "gitsnap [--worktree PATH] [--quiet] <command>") {
		t.Fatalf("usage missing from %q", out)
	}
}

func TestRunUnknownAndFlagErrors(t *testing.T) {
	if err := run(context.Background(), []string{"nope"}); err == nil {
		t.Fatal("expected unknown command error")
	}
	if err := run(context.Background(), []string{"--bad"}); err == nil {
		t.Fatal("expected flag error")
	}
}

func TestRunCommands(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITSNAP_HOME", home)
	worktree := filepath.Join(t.TempDir(), "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, worktree, "a.txt", "hello\n")

	if err := run(context.Background(), []string{"--worktree", worktree, "init"}); err != nil {
		t.Fatal(err)
	}
	saveOut := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--worktree", worktree, "save", "--alias", "first"}); err != nil {
			t.Fatal(err)
		}
	})
	hash := strings.Fields(saveOut)[2]
	if len(hash) != 40 {
		t.Fatalf("hash = %q", hash)
	}
	resolveOut := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--worktree", worktree, "resolve", "first"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.TrimSpace(resolveOut) != hash {
		t.Fatalf("resolve = %q, want %q", strings.TrimSpace(resolveOut), hash)
	}
	aliasesOut := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--worktree", worktree, "aliases"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(aliasesOut, "first "+hash) {
		t.Fatalf("aliases output = %q", aliasesOut)
	}

	write(t, worktree, "a.txt", "world\n")
	write(t, worktree, "b.txt", "new\n")
	filesOut := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--worktree", worktree, "files", "first"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.TrimSpace(filesOut) != "a.txt\nb.txt" {
		t.Fatalf("files = %q", filesOut)
	}
	diffOut := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--worktree", worktree, "diff", "first"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(diffOut, "-hello") || !strings.Contains(diffOut, "+world") {
		t.Fatalf("diff = %q", diffOut)
	}
	if err := run(context.Background(), []string{"--worktree", worktree, "restore", "first", "--", "a.txt"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, worktree, "a.txt"); got != "hello\n" {
		t.Fatalf("a.txt = %q", got)
	}
}

func TestRunCleanup(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITSNAP_HOME", home)
	worktree := filepath.Join(t.TempDir(), "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), []string{"--worktree", worktree, "init"}); err != nil {
		t.Fatal(err)
	}
	ws, err := store.ForWorktree(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ws.RepoDir()); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), []string{"--worktree", worktree, "cleanup"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ws.Root); !os.IsNotExist(err) {
		t.Fatalf("root still exists: %v", err)
	}
}

func TestQuietSuppressesStatus(t *testing.T) {
	t.Setenv("GITSNAP_HOME", filepath.Join(t.TempDir(), "home"))
	worktree := filepath.Join(t.TempDir(), "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, worktree, "a.txt", "hello\n")
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--quiet", "--worktree", worktree, "init"}); err != nil {
			t.Fatal(err)
		}
		if err := run(context.Background(), []string{"--quiet", "--worktree", worktree, "save"}); err != nil {
			t.Fatal(err)
		}
	})
	if out != "" {
		t.Fatalf("out = %q", out)
	}
}

func TestRunUsageErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITSNAP_HOME", home)
	worktree := t.TempDir()
	if err := run(context.Background(), []string{"--worktree", worktree, "init"}); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"cleanup", "extra"},
		{"init", "extra"},
		{"save", "extra"},
		{"resolve"},
		{"resolve", "a", "b"},
		{"diff"},
		{"diff", "a", "b"},
		{"files"},
		{"files", "a", "b"},
		{"restore"},
		{"nope"},
	} {
		full := append([]string{"--worktree", worktree}, args...)
		if err := run(context.Background(), full); err == nil {
			t.Fatalf("expected error for %#v", full)
		}
	}
}

func TestRunEmptyDiffAndFiles(t *testing.T) {
	t.Setenv("GITSNAP_HOME", filepath.Join(t.TempDir(), "home"))
	worktree := filepath.Join(t.TempDir(), "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, worktree, "a.txt", "hello\n")
	if err := run(context.Background(), []string{"--worktree", worktree, "init"}); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), []string{"--worktree", worktree, "save", "--alias", "same"}); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"--worktree", worktree, "diff", "same"}); err != nil {
			t.Fatal(err)
		}
		if err := run(context.Background(), []string{"--worktree", worktree, "files", "same"}); err != nil {
			t.Fatal(err)
		}
	})
	if out != "no changes\nno changed files\n" {
		t.Fatalf("out = %q", out)
	}
}

func TestSaveFlagError(t *testing.T) {
	ws, aliases := commandDeps(t)
	if err := save(context.Background(), gogit.Backend{}, aliases, ws, []string{"--bad"}, false); err == nil {
		t.Fatal("expected error")
	}
}

func TestRestoreWithNoPathSeparator(t *testing.T) {
	ws, aliases := commandDeps(t)
	write(t, ws.Worktree, "a.txt", "hello\n")
	backend := gogit.Backend{}
	h, err := backend.Save(context.Background(), ws.Worktree, ws.RepoDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := aliases.Set("first", h); err != nil {
		t.Fatal(err)
	}
	write(t, ws.Worktree, "a.txt", "world\n")
	if err := restore(context.Background(), backend, aliases, ws, []string{"first", "a.txt"}, false); err != nil {
		t.Fatal(err)
	}
	if got := read(t, ws.Worktree, "a.txt"); got != "hello\n" {
		t.Fatalf("a.txt = %q", got)
	}
}

func TestRunBeforeInitErrors(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("GITSNAP_HOME", home)
	worktree := t.TempDir()
	for _, args := range [][]string{
		{"aliases"},
		{"save"},
		{"resolve", "first"},
		{"diff", "first"},
		{"files", "first"},
		{"restore", "first"},
	} {
		full := append([]string{"--worktree", worktree}, args...)
		err := run(context.Background(), full)
		if err == nil || !strings.Contains(err.Error(), "gitsnap init") {
			t.Fatalf("err for %#v = %v", full, err)
		}
	}
	if entries, err := os.ReadDir(home); err == nil && len(entries) != 0 {
		t.Fatalf("storage was created before init: %#v", entries)
	}
}

func TestListAliasesEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := listAliases(alias.Store{Path: filepath.Join(t.TempDir(), "aliases.json")}, false); err != nil {
			t.Fatal(err)
		}
	})
	if out != "no aliases\n" {
		t.Fatalf("out = %q", out)
	}
}

func TestMainCallsExit(t *testing.T) {
	oldArgs := os.Args
	oldExit := exit
	defer func() { os.Args = oldArgs; exit = oldExit }()
	os.Args = []string{"gitsnap"}
	exit = func(code int) { panic(code) }
	defer func() {
		got := recover()
		if got != 0 {
			t.Fatalf("exit code = %#v", got)
		}
	}()
	captureStdout(t, main)
}

func commandDeps(t *testing.T) (store.WorktreeStore, alias.Store) {
	t.Helper()
	t.Setenv("GITSNAP_HOME", filepath.Join(t.TempDir(), "home"))
	worktree := filepath.Join(t.TempDir(), "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	ws, err := store.ForWorktree(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if err := ws.Ensure(); err != nil {
		t.Fatal(err)
	}
	return ws, alias.Store{Path: ws.AliasPath()}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func write(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root string, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
