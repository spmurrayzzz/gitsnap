package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestForWorktreeUsesGitsnapHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITSNAP_HOME", home)
	worktree := filepath.Join(t.TempDir(), "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}

	store, err := ForWorktree(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(filepath.Dir(store.Root)) != home {
		t.Fatalf("root = %q, want under %q", store.Root, home)
	}
	if store.RepoDir() != filepath.Join(store.Root, "repo") {
		t.Fatalf("repo dir = %q", store.RepoDir())
	}
	if store.AliasPath() != filepath.Join(store.Root, "aliases.json") {
		t.Fatalf("alias path = %q", store.AliasPath())
	}
	if err := store.Ensure(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(store.RepoDir()); err != nil {
		t.Fatal(err)
	}
}
