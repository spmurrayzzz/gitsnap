package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForWorktreeAbsError(t *testing.T) {
	old := absPath
	defer func() { absPath = old }()
	absPath = func(string) (string, error) { return "", errors.New("bad abs") }
	if _, err := ForWorktree("."); err == nil {
		t.Fatal("expected error")
	}
}

func TestForWorktreeRootError(t *testing.T) {
	t.Setenv("GITSNAP_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	old := userHomeDir
	defer func() { userHomeDir = old }()
	userHomeDir = func() (string, error) { return "", errors.New("no home") }
	if _, err := ForWorktree(t.TempDir()); err == nil {
		t.Fatal("expected error")
	}
}

func TestRootUserHomeError(t *testing.T) {
	t.Setenv("GITSNAP_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	old := userHomeDir
	defer func() { userHomeDir = old }()
	userHomeDir = func() (string, error) { return "", errors.New("no home") }
	if _, err := Root(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRootBranches(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("GITSNAP_HOME", "")
	t.Setenv("XDG_DATA_HOME", xdg)
	root, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	if root != filepath.Join(xdg, "gitsnap") {
		t.Fatalf("root = %q", root)
	}

	t.Setenv("XDG_DATA_HOME", "")
	root, err = Root()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(root, filepath.Join(".local", "share", "gitsnap")) {
		t.Fatalf("root = %q", root)
	}
}

func TestForWorktreeEvalSymlinkFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITSNAP_HOME", home)
	missing := filepath.Join(t.TempDir(), "missing")
	store, err := ForWorktree(missing)
	if err != nil {
		t.Fatal(err)
	}
	if store.Worktree != missing {
		t.Fatalf("worktree = %q, want %q", store.Worktree, missing)
	}
}

func TestEnsureError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := WorktreeStore{Root: file}
	if err := store.Ensure(); err == nil {
		t.Fatal("expected ensure error")
	}
}
