package gitsnap

import (
	"context"
	"fmt"

	"github.com/spmurray/gitsnap/internal/alias"
	"github.com/spmurray/gitsnap/internal/backend/gogit"
	"github.com/spmurray/gitsnap/internal/ref"
	"github.com/spmurray/gitsnap/internal/store"
)

type Alias struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	UpdatedAt string `json:"updated_at"`
}

type Client struct {
	Backend gogit.Backend
}

func New() Client {
	return Client{Backend: gogit.Backend{}}
}

func (c Client) Init(ctx context.Context, worktree string) error {
	ws, err := store.ForWorktree(worktree)
	if err != nil {
		return err
	}
	if err := ws.Ensure(); err != nil {
		return err
	}
	return c.Backend.Init(ctx, ws.Worktree, ws.RepoDir())
}

func (c Client) Cleanup(worktree string) error {
	ws, err := store.ForWorktree(worktree)
	if err != nil {
		return err
	}
	return ws.Cleanup()
}

func (c Client) Save(
	ctx context.Context,
	worktree string,
	name string,
) (string, error) {
	ws, aliases, err := c.initialized(worktree)
	if err != nil {
		return "", err
	}
	h, err := c.Backend.Save(ctx, ws.Worktree, ws.RepoDir())
	if err != nil {
		return "", err
	}
	if name != "" {
		if err := aliases.Set(name, h); err != nil {
			return "", err
		}
	}
	return string(h), nil
}

func (c Client) Resolve(worktree string, name string) (string, error) {
	_, aliases, err := c.initialized(worktree)
	if err != nil {
		return "", err
	}
	h, err := ref.Resolve(aliases, name)
	return string(h), err
}

func (c Client) Diff(
	ctx context.Context,
	worktree string,
	name string,
) (string, error) {
	ws, aliases, err := c.initialized(worktree)
	if err != nil {
		return "", err
	}
	h, err := ref.Resolve(aliases, name)
	if err != nil {
		return "", err
	}
	out, err := c.Backend.Diff(ctx, ws.Worktree, ws.RepoDir(), h)
	return string(out), err
}

func (c Client) Files(
	ctx context.Context,
	worktree string,
	name string,
) ([]string, error) {
	ws, aliases, err := c.initialized(worktree)
	if err != nil {
		return nil, err
	}
	h, err := ref.Resolve(aliases, name)
	if err != nil {
		return nil, err
	}
	return c.Backend.ChangedFiles(ctx, ws.Worktree, ws.RepoDir(), h)
}

func (c Client) Restore(
	ctx context.Context,
	worktree string,
	name string,
	paths []string,
) error {
	ws, aliases, err := c.initialized(worktree)
	if err != nil {
		return err
	}
	h, err := ref.Resolve(aliases, name)
	if err != nil {
		return err
	}
	return c.Backend.Restore(ctx, ws.Worktree, ws.RepoDir(), h, paths)
}

func (c Client) Aliases(worktree string) ([]Alias, error) {
	_, aliases, err := c.initialized(worktree)
	if err != nil {
		return nil, err
	}
	names, records, err := aliases.List()
	if err != nil {
		return nil, err
	}
	out := make([]Alias, 0, len(names))
	for _, name := range names {
		rec := records[name]
		out = append(out, Alias{
			Name:      name,
			Hash:      string(rec.Hash),
			UpdatedAt: rec.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return out, nil
}

func (c Client) initialized(
	worktree string,
) (store.WorktreeStore, alias.Store, error) {
	ws, err := store.ForWorktree(worktree)
	if err != nil {
		return store.WorktreeStore{}, alias.Store{}, err
	}
	ok, err := ws.Initialized()
	if err != nil {
		return store.WorktreeStore{}, alias.Store{}, err
	}
	if !ok {
		return store.WorktreeStore{}, alias.Store{}, fmt.Errorf(
			"worktree has not been initialized; run gitsnap init",
		)
	}
	return ws, alias.Store{Path: ws.AliasPath()}, nil
}
