package gogit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"github.com/spmurray/gitsnap/internal/treehash"
)

type Backend struct{}

var buildTreeFunc = buildTree
var getTreeFunc = object.GetTree
var patchFunc = defaultPatch
var diffFunc = defaultDiff
var setIndexFunc = defaultSetIndex
var statusFunc = defaultStatus
var addFunc = defaultAdd
var indexFunc = defaultIndex
var gitOpenFunc = git.Open
var gitInitFunc = git.Init
var worktreeFunc = defaultWorktree
var fileSetFunc = fileSet
var fileReaderFunc = defaultFileReader
var copyFunc = io.Copy
var treeFunc = defaultTree
var lstatFunc = os.Lstat
var removeAllFunc = os.RemoveAll
var symlinkFunc = os.Symlink
var writeDiskFileFunc = os.WriteFile

func defaultPatch(
	ctx context.Context,
	from *object.Tree,
	to *object.Tree,
) (*object.Patch, error) {
	return from.PatchContext(ctx, to)
}

func defaultDiff(
	ctx context.Context,
	from *object.Tree,
	to *object.Tree,
) (object.Changes, error) {
	return from.DiffContext(ctx, to)
}

func defaultSetIndex(r *git.Repository, idx *index.Index) error {
	return r.Storer.SetIndex(idx)
}

func defaultStatus(wt *git.Worktree) (git.Status, error) {
	return wt.Status()
}

func defaultAdd(wt *git.Worktree, p string) error {
	return wt.AddWithOptions(&git.AddOptions{Path: p})
}

func defaultIndex(r *git.Repository) (*index.Index, error) {
	return r.Storer.Index()
}

func defaultWorktree(r *git.Repository) (*git.Worktree, error) {
	return r.Worktree()
}

func defaultFileReader(file *object.File) (io.ReadCloser, error) {
	return file.Reader()
}

func defaultTree(root *object.Tree, p string) (*object.Tree, error) {
	return root.Tree(p)
}

func (Backend) Init(ctx context.Context, worktree string, store string) error {
	_, _, err := openRepo(worktree, store)
	return err
}

func (Backend) Save(
	ctx context.Context,
	worktree string,
	store string,
) (treehash.Hash, error) {
	r, wt, err := openRepo(worktree, store)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	wt.Excludes = append(wt.Excludes, gitignore.ParsePattern(".git/", nil))
	if err := setIndexFunc(r, &index.Index{Version: 2}); err != nil {
		return "", err
	}
	status, err := statusFunc(wt)
	if err != nil {
		return "", err
	}
	paths := make([]string, 0, len(status))
	for p := range status {
		p = filepath.ToSlash(p)
		if p == ".git" || strings.HasPrefix(p, ".git/") {
			continue
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		if err := addFunc(wt, p); err != nil {
			return "", err
		}
	}
	idx, err := indexFunc(r)
	if err != nil {
		return "", err
	}
	h, err := buildTreeFunc(r.Storer, idx)
	if err != nil {
		return "", err
	}
	return treehash.Hash(h.String()), nil
}

func (b Backend) Diff(
	ctx context.Context,
	worktree string,
	store string,
	base treehash.Hash,
) ([]byte, error) {
	r, err := openOnly(worktree, store)
	if err != nil {
		return nil, err
	}
	current, err := b.Save(ctx, worktree, store)
	if err != nil {
		return nil, err
	}
	from, err := getTreeFunc(r.Storer, plumbing.NewHash(string(base)))
	if err != nil {
		return nil, err
	}
	to, err := getTreeFunc(r.Storer, plumbing.NewHash(string(current)))
	if err != nil {
		return nil, err
	}
	patch, err := patchFunc(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return []byte(patch.String()), nil
}

func (b Backend) ChangedFiles(
	ctx context.Context,
	worktree string,
	store string,
	base treehash.Hash,
) ([]string, error) {
	r, err := openOnly(worktree, store)
	if err != nil {
		return nil, err
	}
	current, err := b.Save(ctx, worktree, store)
	if err != nil {
		return nil, err
	}
	from, err := getTreeFunc(r.Storer, plumbing.NewHash(string(base)))
	if err != nil {
		return nil, err
	}
	to, err := getTreeFunc(r.Storer, plumbing.NewHash(string(current)))
	if err != nil {
		return nil, err
	}
	changes, err := diffFunc(ctx, from, to)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var files []string
	for _, change := range changes {
		a, b, err := change.Files()
		if err != nil {
			return nil, err
		}
		if a != nil && !seen[change.From.Name] {
			seen[change.From.Name] = true
			files = append(files, change.From.Name)
		}
		if b != nil && !seen[change.To.Name] {
			seen[change.To.Name] = true
			files = append(files, change.To.Name)
		}
	}
	sort.Strings(files)
	return files, nil
}

func (b Backend) Restore(
	ctx context.Context,
	worktree string,
	store string,
	tree treehash.Hash,
	paths []string,
) error {
	r, err := openOnly(worktree, store)
	if err != nil {
		return err
	}
	root, err := getTreeFunc(r.Storer, plumbing.NewHash(string(tree)))
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		current, err := b.Save(ctx, worktree, store)
		if err != nil {
			return err
		}
		currentRoot, err := getTreeFunc(r.Storer, plumbing.NewHash(string(current)))
		if err != nil {
			return err
		}
		if err := removeFilesNotIn(ctx, currentRoot, root, worktree); err != nil {
			return err
		}
		return restoreTree(ctx, root, worktree, "")
	}
	for _, p := range paths {
		if err := restorePath(ctx, root, worktree, p); err != nil {
			return err
		}
	}
	return nil
}

func openRepo(worktree string, store string) (*git.Repository, *git.Worktree, error) {
	if err := os.MkdirAll(store, 0o755); err != nil {
		return nil, nil, err
	}
	s := filesystem.NewStorage(osfs.New(store), cache.NewObjectLRUDefault())
	wtfs := hideGitFS{Filesystem: osfs.New(worktree)}
	r, err := gitOpenFunc(s, wtfs)
	if err == git.ErrRepositoryNotExists {
		if _, err := gitInitFunc(s, nil); err != nil {
			return nil, nil, err
		}
		r, err = gitOpenFunc(s, wtfs)
	}
	if err != nil {
		return nil, nil, err
	}
	wt, err := worktreeFunc(r)
	if err != nil {
		return nil, nil, err
	}
	return r, wt, nil
}

func openOnly(worktree string, store string) (*git.Repository, error) {
	r, _, err := openRepo(worktree, store)
	return r, err
}

type hideGitFS struct {
	billy.Filesystem
}

func (fs hideGitFS) Open(name string) (billy.File, error) {
	if isGitPath(name) {
		return nil, os.ErrNotExist
	}
	return fs.Filesystem.Open(name)
}

func (fs hideGitFS) OpenFile(
	name string,
	flag int,
	perm os.FileMode,
) (billy.File, error) {
	if isGitPath(name) {
		return nil, os.ErrNotExist
	}
	return fs.Filesystem.OpenFile(name, flag, perm)
}

func (fs hideGitFS) Stat(name string) (os.FileInfo, error) {
	if isGitPath(name) {
		return nil, os.ErrNotExist
	}
	return fs.Filesystem.Stat(name)
}

func (fs hideGitFS) Lstat(name string) (os.FileInfo, error) {
	if isGitPath(name) {
		return nil, os.ErrNotExist
	}
	return fs.Filesystem.Lstat(name)
}

func (fs hideGitFS) ReadDir(name string) ([]os.FileInfo, error) {
	infos, err := fs.Filesystem.ReadDir(name)
	if err != nil {
		return nil, err
	}
	out := infos[:0]
	for _, info := range infos {
		if info.Name() != ".git" {
			out = append(out, info)
		}
	}
	return out, nil
}

func (fs hideGitFS) Chroot(name string) (billy.Filesystem, error) {
	if isGitPath(name) {
		return nil, os.ErrNotExist
	}
	chroot, err := fs.Filesystem.Chroot(name)
	if err != nil {
		return nil, err
	}
	return hideGitFS{Filesystem: chroot}, nil
}

func isGitPath(name string) bool {
	name = filepath.ToSlash(filepath.Clean(name))
	for _, part := range strings.Split(name, "/") {
		if part == ".git" {
			return true
		}
	}
	return false
}

func buildTree(s storage.Storer, idx *index.Index) (plumbing.Hash, error) {
	builder := treeBuilder{store: s, trees: map[string]*object.Tree{"": {}}}
	for _, e := range idx.Entries {
		builder.add(e)
	}
	return builder.write("", builder.trees[""])
}

type treeBuilder struct {
	store storage.Storer
	trees map[string]*object.Tree
}

func (b *treeBuilder) add(e *index.Entry) {
	parts := strings.Split(e.Name, "/")
	var full string
	for _, part := range parts {
		parent := full
		full = path.Join(full, part)
		if b.hasEntry(parent, path.Base(full)) {
			continue
		}
		entry := object.TreeEntry{Name: path.Base(full)}
		if full == e.Name {
			entry.Mode = e.Mode
			entry.Hash = e.Hash
		} else {
			entry.Mode = filemode.Dir
			b.trees[full] = &object.Tree{}
		}
		b.trees[parent].Entries = append(b.trees[parent].Entries, entry)
	}
}

func (b *treeBuilder) hasEntry(parent string, name string) bool {
	for _, e := range b.trees[parent].Entries {
		if e.Name == name {
			return true
		}
	}
	return false
}

func (b *treeBuilder) write(parent string, t *object.Tree) (plumbing.Hash, error) {
	sort.Sort(entries(t.Entries))
	for i, e := range t.Entries {
		if e.Mode != filemode.Dir {
			continue
		}
		h, err := b.write(path.Join(parent, e.Name), b.trees[path.Join(parent, e.Name)])
		if err != nil {
			return plumbing.ZeroHash, err
		}
		e.Hash = h
		t.Entries[i] = e
	}
	o := b.store.NewEncodedObject()
	if err := t.Encode(o); err != nil {
		return plumbing.ZeroHash, err
	}
	return b.store.SetEncodedObject(o)
}

type entries []object.TreeEntry

func (e entries) Len() int      { return len(e) }
func (e entries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e entries) Less(i, j int) bool {
	return sortName(e[i]) < sortName(e[j])
}

func sortName(e object.TreeEntry) string {
	if e.Mode == filemode.Dir {
		return e.Name + "/"
	}
	return e.Name
}

func removeFilesNotIn(
	ctx context.Context,
	current *object.Tree,
	target *object.Tree,
	worktree string,
) error {
	targetFiles, err := fileSetFunc(target)
	if err != nil {
		return err
	}
	return current.Files().ForEach(func(file *object.File) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if targetFiles[file.Name] {
			return nil
		}
		return os.Remove(filepath.Join(worktree, filepath.FromSlash(file.Name)))
	})
}

func fileSet(tree *object.Tree) (map[string]bool, error) {
	files := map[string]bool{}
	err := tree.Files().ForEach(func(file *object.File) error {
		files[file.Name] = true
		return nil
	})
	return files, err
}

func restorePath(ctx context.Context, root *object.Tree, worktree string, p string) error {
	p = filepath.ToSlash(filepath.Clean(p))
	entry, err := root.FindEntry(p)
	if err != nil {
		return err
	}
	if entry.Mode == filemode.Dir {
		t, err := treeFunc(root, p)
		if err != nil {
			return err
		}
		return restoreTree(ctx, t, worktree, p)
	}
	return restoreFile(root, worktree, p)
}

func restoreTree(ctx context.Context, tree *object.Tree, worktree string, prefix string) error {
	return tree.Files().ForEach(func(file *object.File) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := file.Name
		if prefix != "" {
			name = path.Join(prefix, file.Name)
		}
		return writeFile(worktree, name, file)
	})
}

func restoreFile(root *object.Tree, worktree string, p string) error {
	file, err := root.File(p)
	if err != nil {
		return err
	}
	return writeFile(worktree, p, file)
}

func writeFile(worktree string, name string, file *object.File) error {
	if strings.HasPrefix(name, "../") || path.IsAbs(name) {
		return fmt.Errorf("invalid path %q", name)
	}
	target := filepath.Join(worktree, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	reader, err := fileReaderFunc(file)
	if err != nil {
		return err
	}
	defer reader.Close()
	var buf bytes.Buffer
	if _, err := copyFunc(&buf, reader); err != nil {
		return err
	}
	if file.Mode == filemode.Symlink {
		if err := removeTarget(target); err != nil {
			return err
		}
		return symlinkFunc(string(buf.Bytes()), target)
	}
	if err := removeTargetIfConflict(target); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if file.Mode == filemode.Executable {
		mode = 0o755
	}
	return writeDiskFileFunc(target, buf.Bytes(), mode)
}

func removeTarget(target string) error {
	if _, err := lstatFunc(target); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return removeAllFunc(target)
}

func removeTargetIfConflict(target string) error {
	info, err := lstatFunc(target)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return removeAllFunc(target)
	}
	return nil
}
