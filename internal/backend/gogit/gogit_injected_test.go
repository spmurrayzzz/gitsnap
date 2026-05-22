package gogit

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage"
)

func TestInjectedOpenRepoErrorBranches(t *testing.T) {
	worktree, store := dirs(t)
	gitInitFunc = func(storage.Storer, billy.Filesystem) (*git.Repository, error) {
		return nil, errors.New("init")
	}
	if _, _, err := openRepo(worktree, store); err == nil {
		t.Fatal("expected init error")
	}
	resetInjected()

	calls := 0
	gitOpenFunc = func(storage.Storer, billy.Filesystem) (*git.Repository, error) {
		calls++
		if calls == 1 {
			return nil, git.ErrRepositoryNotExists
		}
		return nil, errors.New("open")
	}
	if _, _, err := openRepo(worktree, store); err == nil {
		t.Fatal("expected second open error")
	}
	resetInjected()

	worktreeFunc = func(*git.Repository) (*git.Worktree, error) {
		return nil, errors.New("worktree")
	}
	if _, _, err := openRepo(worktree, store); err == nil {
		t.Fatal("expected worktree error")
	}
	resetInjected()
}

func TestInjectedErrorBranches(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	statusFunc = func(*git.Worktree) (git.Status, error) {
		return git.Status{
			".git":        &git.FileStatus{},
			".git/config": &git.FileStatus{},
		}, nil
	}
	if _, err := backend.Save(context.Background(), worktree, store); err != nil {
		t.Fatal(err)
	}
	resetInjected()

	setIndexFunc = func(*git.Repository, *index.Index) error {
		return errors.New("set index")
	}
	if _, err := backend.Save(context.Background(), worktree, store); err == nil {
		t.Fatal("expected set index error")
	}
	resetInjected()

	statusFunc = func(*git.Worktree) (git.Status, error) {
		return nil, errors.New("status")
	}
	if _, err := backend.Save(context.Background(), worktree, store); err == nil {
		t.Fatal("expected status error")
	}
	resetInjected()

	addFunc = func(*git.Worktree, string) error {
		return errors.New("add")
	}
	if _, err := backend.Save(context.Background(), worktree, store); err == nil {
		t.Fatal("expected add error")
	}
	resetInjected()

	indexFunc = func(*git.Repository) (*index.Index, error) {
		return nil, errors.New("index")
	}
	if _, err := backend.Save(context.Background(), worktree, store); err == nil {
		t.Fatal("expected index error")
	}
	resetInjected()

	buildTreeFunc = func(storage.Storer, *index.Index) (plumbing.Hash, error) {
		return plumbing.ZeroHash, errors.New("build")
	}
	if _, err := backend.Save(context.Background(), worktree, store); err == nil {
		t.Fatal("expected build tree error")
	}
	resetInjected()

	calls := 0
	getTreeFunc = func(s storer.EncodedObjectStorer, h plumbing.Hash) (*object.Tree, error) {
		calls++
		if calls == 2 {
			return nil, errors.New("to tree")
		}
		return object.GetTree(s, h)
	}
	if _, err := backend.Diff(context.Background(), worktree, store, base); err == nil {
		t.Fatal("expected to tree error")
	}
	resetInjected()

	patchFunc = func(context.Context, *object.Tree, *object.Tree) (*object.Patch, error) {
		return nil, errors.New("patch")
	}
	if _, err := backend.Diff(context.Background(), worktree, store, base); err == nil {
		t.Fatal("expected patch error")
	}
	resetInjected()

	calls = 0
	getTreeFunc = func(s storer.EncodedObjectStorer, h plumbing.Hash) (*object.Tree, error) {
		calls++
		if calls == 2 {
			return nil, errors.New("to tree")
		}
		return object.GetTree(s, h)
	}
	if _, err := backend.ChangedFiles(context.Background(), worktree, store, base); err == nil {
		t.Fatal("expected files to tree error")
	}
	resetInjected()

	diffFunc = func(context.Context, *object.Tree, *object.Tree) (object.Changes, error) {
		return nil, errors.New("diff")
	}
	if _, err := backend.ChangedFiles(context.Background(), worktree, store, base); err == nil {
		t.Fatal("expected diff error")
	}
	resetInjected()

	diffFunc = func(context.Context, *object.Tree, *object.Tree) (object.Changes, error) {
		return object.Changes{&object.Change{}}, nil
	}
	if _, err := backend.ChangedFiles(context.Background(), worktree, store, base); err == nil {
		t.Fatal("expected change files error")
	}
	resetInjected()

	calls = 0
	getTreeFunc = func(s storer.EncodedObjectStorer, h plumbing.Hash) (*object.Tree, error) {
		calls++
		if calls == 2 {
			return nil, errors.New("current tree")
		}
		return object.GetTree(s, h)
	}
	if err := backend.Restore(context.Background(), worktree, store, base, nil); err == nil {
		t.Fatal("expected current tree error")
	}
	resetInjected()

	fileSetFunc = func(*object.Tree) (map[string]bool, error) {
		return nil, errors.New("fileset")
	}
	if err := backend.Restore(context.Background(), worktree, store, base, nil); err == nil {
		t.Fatal("expected remove files error")
	}
	resetInjected()
}

func TestInjectedHelperErrors(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	write(t, worktree, "dir/a.txt", "hello\n")
	if err := os.Symlink("a.txt", filepath.Join(worktree, "link")); err != nil {
		t.Fatal(err)
	}
	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	r, err := openOnly(worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	root, err := object.GetTree(r.Storer, plumbing.NewHash(string(base)))
	if err != nil {
		t.Fatal(err)
	}

	fileSetFunc = func(*object.Tree) (map[string]bool, error) {
		return nil, errors.New("fileset")
	}
	if err := removeFilesNotIn(context.Background(), root, root, worktree); err == nil {
		t.Fatal("expected fileset error")
	}
	resetInjected()

	file, err := root.File("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := removeTarget(filepath.Join(worktree, "missing")); err != nil {
		t.Fatal(err)
	}
	lstatFunc = func(string) (os.FileInfo, error) {
		return nil, errors.New("lstat")
	}
	if err := removeTarget(filepath.Join(worktree, "a.txt")); err == nil {
		t.Fatal("expected remove target lstat error")
	}
	if err := removeTargetIfConflict(filepath.Join(worktree, "a.txt")); err == nil {
		t.Fatal("expected conflict lstat error")
	}
	if err := writeFile(worktree, "lstat.txt", file); err == nil {
		t.Fatal("expected conflict error")
	}
	resetInjected()
	if err := removeTargetIfConflict(filepath.Join(worktree, "missing")); err != nil {
		t.Fatal(err)
	}

	fileReaderFunc = func(*object.File) (io.ReadCloser, error) {
		return nil, errors.New("reader")
	}
	if err := writeFile(worktree, "reader.txt", file); err == nil {
		t.Fatal("expected reader error")
	}
	resetInjected()

	symlinkFile, err := root.File("link")
	if err != nil {
		t.Fatal(err)
	}
	write(t, worktree, "existing", "x")
	removeAllFunc = func(string) error {
		return errors.New("remove")
	}
	if err := writeFile(worktree, "existing", symlinkFile); err == nil {
		t.Fatal("expected symlink remove error")
	}
	resetInjected()

	copyFunc = func(io.Writer, io.Reader) (int64, error) {
		return 0, errors.New("copy")
	}
	if err := writeFile(worktree, "copy.txt", file); err == nil {
		t.Fatal("expected copy error")
	}
	resetInjected()

	treeFunc = func(*object.Tree, string) (*object.Tree, error) {
		return nil, errors.New("tree")
	}
	if err := restorePath(context.Background(), root, worktree, "dir"); err == nil {
		t.Fatal("expected tree error")
	}
	resetInjected()
}

func TestDefaultInjectedFuncs(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	r, err := openOnly(worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := object.GetTree(r.Storer, plumbing.NewHash(string(base)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := patchFunc(context.Background(), tree, tree); err != nil {
		t.Fatal(err)
	}
	if _, err := diffFunc(context.Background(), tree, tree); err != nil {
		t.Fatal(err)
	}
}

func resetInjected() {
	buildTreeFunc = buildTree
	getTreeFunc = object.GetTree
	patchFunc = defaultPatch
	diffFunc = defaultDiff
	setIndexFunc = defaultSetIndex
	statusFunc = defaultStatus
	addFunc = defaultAdd
	indexFunc = defaultIndex
	gitOpenFunc = git.Open
	gitInitFunc = git.Init
	worktreeFunc = defaultWorktree
	fileSetFunc = fileSet
	fileReaderFunc = defaultFileReader
	copyFunc = io.Copy
	treeFunc = defaultTree
	lstatFunc = os.Lstat
	removeAllFunc = os.RemoveAll
	symlinkFunc = os.Symlink
	writeDiskFileFunc = os.WriteFile
}
