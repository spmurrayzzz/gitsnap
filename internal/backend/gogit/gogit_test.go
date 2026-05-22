package gogit

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestSaveRespectsGitignoreAndExcludesDotGit(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, ".gitignore", "*.log\n")
	write(t, worktree, "a.txt", "hello\n")
	write(t, worktree, "ignored.log", "one\n")
	write(t, worktree, ".git/config", "one\n")

	backend := Backend{}
	first, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	write(t, worktree, "ignored.log", "two\n")
	write(t, worktree, ".git/config", "two\n")

	second, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Fatalf("hash changed after ignored files changed: %s != %s", first, second)
	}
}

func TestChangedFilesAndDiff(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")

	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	write(t, worktree, "a.txt", "world\n")
	write(t, worktree, "b.txt", "new\n")

	files, err := backend.ChangedFiles(context.Background(), worktree, store, base)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.txt", "b.txt"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %#v, want %#v", files, want)
	}

	diff, err := backend.Diff(context.Background(), worktree, store, base)
	if err != nil {
		t.Fatal(err)
	}
	text := string(diff)
	for _, want := range []string{"diff --git a/a.txt b/a.txt", "-hello", "+world", "diff --git a/b.txt b/b.txt"} {
		if !strings.Contains(text, want) {
			t.Fatalf("diff missing %q:\n%s", want, text)
		}
	}
}

func TestRestoreFull(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, ".gitignore", "ignored.log\n")
	write(t, worktree, "a.txt", "hello\n")
	write(t, worktree, "ignored.log", "keep\n")

	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	write(t, worktree, "a.txt", "world\n")
	write(t, worktree, "b.txt", "remove\n")
	write(t, worktree, "ignored.log", "kept\n")

	if err := backend.Restore(context.Background(), worktree, store, base, nil); err != nil {
		t.Fatal(err)
	}
	if got := read(t, worktree, "a.txt"); got != "hello\n" {
		t.Fatalf("a.txt = %q", got)
	}
	if _, err := os.Stat(filepath.Join(worktree, "b.txt")); !os.IsNotExist(err) {
		t.Fatalf("b.txt still exists or stat failed: %v", err)
	}
	if got := read(t, worktree, "ignored.log"); got != "kept\n" {
		t.Fatalf("ignored.log = %q", got)
	}
}

func TestRestorePath(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "hello\n")
	write(t, worktree, "b.txt", "first\n")

	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	write(t, worktree, "a.txt", "world\n")
	write(t, worktree, "b.txt", "second\n")

	if err := backend.Restore(context.Background(), worktree, store, base, []string{"a.txt"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, worktree, "a.txt"); got != "hello\n" {
		t.Fatalf("a.txt = %q", got)
	}
	if got := read(t, worktree, "b.txt"); got != "second\n" {
		t.Fatalf("b.txt = %q", got)
	}
}

func TestRestoreRegularFileReplacesSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on windows")
	}
	worktree, store := dirs(t)
	write(t, worktree, "a.txt", "snapshot\n")

	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(worktree, "a.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(worktree, "a.txt")); err != nil {
		t.Fatal(err)
	}

	if err := backend.Restore(context.Background(), worktree, store, base, []string{"a.txt"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, worktree, "a.txt"); got != "snapshot\n" {
		t.Fatalf("a.txt = %q", got)
	}
	if got := readFile(t, outside); got != "outside\n" {
		t.Fatalf("outside = %q", got)
	}
	if _, err := os.Readlink(filepath.Join(worktree, "a.txt")); err == nil {
		t.Fatal("a.txt is still a symlink")
	}
}

func TestRestoreFileDirectoryConflict(t *testing.T) {
	worktree, store := dirs(t)
	write(t, worktree, "d", "snapshot\n")

	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(worktree, "d")); err != nil {
		t.Fatal(err)
	}
	write(t, worktree, "d/a.txt", "current\n")

	if err := backend.Restore(context.Background(), worktree, store, base, nil); err != nil {
		t.Fatal(err)
	}
	if got := read(t, worktree, "d"); got != "snapshot\n" {
		t.Fatalf("d = %q", got)
	}
}

func TestRestoreSymlinkAndExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on windows")
	}
	worktree, store := dirs(t)
	write(t, worktree, "script.sh", "#!/bin/sh\n")
	if err := os.Chmod(filepath.Join(worktree, "script.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("script.sh", filepath.Join(worktree, "link")); err != nil {
		t.Fatal(err)
	}

	backend := Backend{}
	base, err := backend.Save(context.Background(), worktree, store)
	if err != nil {
		t.Fatal(err)
	}

	write(t, worktree, "script.sh", "changed\n")
	if err := os.Remove(filepath.Join(worktree, "link")); err != nil {
		t.Fatal(err)
	}
	write(t, worktree, "link", "not a link\n")

	if err := backend.Restore(context.Background(), worktree, store, base, nil); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(worktree, "script.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("script.sh is not executable: %v", info.Mode())
	}
	target, err := os.Readlink(filepath.Join(worktree, "link"))
	if err != nil {
		t.Fatal(err)
	}
	if target != "script.sh" {
		t.Fatalf("link target = %q", target)
	}
}

func dirs(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	worktree := filepath.Join(root, "worktree")
	store := filepath.Join(root, "store")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	return worktree, store
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
	return readFile(t, filepath.Join(root, filepath.FromSlash(name)))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
