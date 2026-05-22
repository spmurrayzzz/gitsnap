package store

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

type WorktreeStore struct {
	Root     string
	Worktree string
	ID       string
}

var absPath = filepath.Abs
var userHomeDir = os.UserHomeDir

func ForWorktree(worktree string) (WorktreeStore, error) {
	abs, err := absPath(worktree)
	if err != nil {
		return WorktreeStore{}, err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		real = abs
	}
	root, err := Root()
	if err != nil {
		return WorktreeStore{}, err
	}
	sum := sha256.Sum256([]byte(real))
	id := hex.EncodeToString(sum[:])
	return WorktreeStore{Root: filepath.Join(root, "worktrees", id), Worktree: real, ID: id}, nil
}

func Root() (string, error) {
	if v := os.Getenv("GITSNAP_HOME"); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return filepath.Join(v, "gitsnap"), nil
	}
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "gitsnap"), nil
}

func (s WorktreeStore) RepoDir() string {
	return filepath.Join(s.Root, "repo")
}

func (s WorktreeStore) AliasPath() string {
	return filepath.Join(s.Root, "aliases.json")
}

func (s WorktreeStore) Ensure() error {
	return os.MkdirAll(s.RepoDir(), 0o755)
}
