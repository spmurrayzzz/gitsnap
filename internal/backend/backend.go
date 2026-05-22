package backend

import (
	"context"

	"github.com/spmurray/gitsnap/internal/treehash"
)

type Backend interface {
	Init(ctx context.Context, worktree string, store string) error
	Save(ctx context.Context, worktree string, store string) (treehash.Hash, error)
	Diff(
		ctx context.Context,
		worktree string,
		store string,
		base treehash.Hash,
	) ([]byte, error)
	ChangedFiles(
		ctx context.Context,
		worktree string,
		store string,
		base treehash.Hash,
	) ([]string, error)
	Restore(
		ctx context.Context,
		worktree string,
		store string,
		tree treehash.Hash,
		paths []string,
	) error
}
