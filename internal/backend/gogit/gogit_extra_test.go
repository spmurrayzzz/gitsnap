package gogit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spmurray/gitsnap/internal/treehash"
)

func TestInitAndErrorBranches(t *testing.T) {
	worktree, store := dirs(t)
	backend := Backend{}
	if err := backend.Init(context.Background(), worktree, store); err != nil {
		t.Fatal(err)
	}
	badStore := filepath.Join(t.TempDir(), "store")
	if err := os.WriteFile(badStore, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := backend.Init(context.Background(), worktree, badStore); err == nil {
		t.Fatal("expected init error")
	}
	if _, err := backend.Save(context.Background(), worktree, badStore); err == nil {
		t.Fatal("expected save open error")
	}
}

func TestSaveSetIndexError(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	backend := Backend{}
	if err := backend.Init(context.Background(), worktree, store); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(store, "index"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Save(context.Background(), worktree, store); err == nil {
		t.Fatal("expected set index error")
	}
}

func TestSaveCanceled(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Backend{}).Save(ctx, worktree, store); err == nil {
		t.Fatal("expected canceled error")
	}
}

func TestOpenOnlyOperationErrors(t *testing.T) {
	worktree, _ := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	backend := Backend{}
	badStore := filepath.Join(t.TempDir(), "store")
	if err := os.WriteFile(badStore, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Diff(context.Background(), worktree, badStore, ""); err == nil {
		t.Fatal("expected diff open error")
	}
	if _, err := backend.ChangedFiles(context.Background(), worktree, badStore, ""); err == nil {
		t.Fatal("expected files open error")
	}
	if err := backend.Restore(context.Background(), worktree, badStore, "", nil); err == nil {
		t.Fatal("expected restore open error")
	}
}

func TestDiffAndFilesErrors(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	missing := treehash.Hash("0123456789abcdef0123456789abcdef01234567")
	if _, err := backend.Diff(context.Background(), worktree, store, missing); err == nil {
		t.Fatal("expected diff base error")
	}
	if _, err := backend.ChangedFiles(context.Background(), worktree, store, missing); err == nil {
		t.Fatal("expected files base error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := backend.Diff(ctx, worktree, store, base); err == nil {
		t.Fatal("expected diff canceled error")
	}
	if _, err := backend.ChangedFiles(ctx, worktree, store, base); err == nil {
		t.Fatal("expected files canceled error")
	}
}

func TestRestoreHelperContextErrors(t *testing.T) {
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
	root, err := object.GetTree(r.Storer, plumbing.NewHash(string(base)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := removeFilesNotIn(ctx, root, &object.Tree{}, worktree); err == nil {
		t.Fatal("expected remove context error")
	}
	if err := restoreTree(ctx, root, worktree, ""); err == nil {
		t.Fatal("expected restore tree context error")
	}
}

func TestRestoreErrorsAndDirectoryPath(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "dir/a.txt", "hello\n")
	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	missing := treehash.Hash("0123456789abcdef0123456789abcdef01234567")
	if err := backend.Restore(context.Background(), worktree, store, missing, nil); err == nil {
		t.Fatal("expected missing tree error")
	}
	if err := backend.Restore(context.Background(), worktree, store, base, []string{"missing"}); err == nil {
		t.Fatal("expected missing path error")
	}
	write(t, worktree, "dir/a.txt", "world\n")
	if err := backend.Restore(context.Background(), worktree, store, base, []string{"dir"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, worktree, "dir/a.txt"); got != "hello\n" {
		t.Fatalf("dir/a.txt = %q", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := backend.Restore(ctx, worktree, store, base, nil); err == nil {
		t.Fatal("expected canceled full restore error")
	}
}

func TestHideGitFS(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.txt", "hello\n")
	write(t, root, ".git/config", "secret\n")
	fs := hideGitFS{Filesystem: osfs.New(root)}
	if _, err := fs.Open("a.txt"); err != nil {
		t.Fatal(err)
	}
	f, err := fs.OpenFile("a.txt", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if _, err := fs.Open(".git/config"); err == nil {
		t.Fatal("expected open .git error")
	}
	if _, err := fs.OpenFile(".git/config", os.O_RDONLY, 0); err == nil {
		t.Fatal("expected openfile .git error")
	}
	if _, err := fs.Stat("a.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Stat(".git/config"); err == nil {
		t.Fatal("expected stat .git error")
	}
	if _, err := fs.Lstat("a.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Lstat(".git/config"); err == nil {
		t.Fatal("expected lstat .git error")
	}
	infos, err := fs.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, info := range infos {
		names = append(names, info.Name())
	}
	if !reflect.DeepEqual(names, []string{"a.txt"}) {
		t.Fatalf("names = %#v", names)
	}
	if _, err := fs.ReadDir("missing"); err == nil {
		t.Fatal("expected readdir error")
	}
	if _, err := fs.Chroot("."); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Chroot(".git"); err == nil {
		t.Fatal("expected chroot .git error")
	}
	errFS := hideGitFS{Filesystem: chrootErrorFS{Filesystem: osfs.New(root)}}
	if _, err := errFS.Chroot("."); err == nil {
		t.Fatal("expected chroot error")
	}
}

type chrootErrorFS struct {
	billy.Filesystem
}

func (fs chrootErrorFS) Chroot(string) (billy.Filesystem, error) {
	return nil, errors.New("chroot")
}

func TestTreeBuilderWriteErrors(t *testing.T) {
	worktree, store := dirs(t)
	r, _, err := openRepo(worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	builder := treeBuilder{store: r.Storer, trees: map[string]*object.Tree{"": {}}}
	_, err = builder.write("", &object.Tree{Entries: []object.TreeEntry{{Name: "bad\x00"}}})
	if err == nil {
		t.Fatal("expected encode error")
	}
	builder = treeBuilder{store: r.Storer, trees: map[string]*object.Tree{
		"":    {Entries: []object.TreeEntry{{Name: "dir", Mode: filemode.Dir}}},
		"dir": {Entries: []object.TreeEntry{{Name: "bad\x00"}}},
	}}
	_, err = builder.write("", builder.trees[""])
	if err == nil {
		t.Fatal("expected recursive error")
	}
}

func TestTreeBuilderAndHelpers(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "dir/a.txt", "hello\n")
	write(t, worktree, "z.txt", "z\n")
	r, _, err := openRepo(worktree, store)
	if err != nil {
		t.Fatal(err)
	}
	idx := &index.Index{Version: 2}
	idx.Entries = append(idx.Entries,
		&index.Entry{Name: "dir/a.txt", Mode: filemode.Regular},
		&index.Entry{Name: "dir/b.txt", Mode: filemode.Regular},
	)
	builder := treeBuilder{store: r.Storer, trees: map[string]*object.Tree{"": {}}}
	for _, e := range idx.Entries {
		builder.add(e)
	}
	if !builder.hasEntry("", "dir") {
		t.Fatal("expected dir entry")
	}
	if builder.hasEntry("", "missing") {
		t.Fatal("unexpected entry")
	}
	entries := entries{{Name: "b"}, {Name: "a"}}
	entries.Swap(0, 1)
	if entries[0].Name != "a" {
		t.Fatalf("swap failed: %#v", entries)
	}
	if sortName(object.TreeEntry{Name: "dir", Mode: filemode.Dir}) != "dir/" {
		t.Fatal("dir sort name")
	}
}

func TestRestoreFileAndWriteFileErrors(t *testing.T) {
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
	root, err := object.GetTree(r.Storer, plumbing.NewHash(string(base)))
	if err != nil {
		t.Fatal(err)
	}
	file, err := root.File("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFile(t.TempDir(), "../bad", file); err == nil {
		t.Fatal("expected invalid path error")
	}
	badRoot := t.TempDir()
	write(t, badRoot, "parent", "file\n")
	if err := writeFile(badRoot, "parent/a.txt", file); err == nil {
		t.Fatal("expected mkdir error")
	}
	if err := restoreFile(root, worktree, "missing"); err == nil {
		t.Fatal("expected restore file error")
	}
}
