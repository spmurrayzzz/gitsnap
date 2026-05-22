package alias

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spmurray/gitsnap/internal/treehash"
)

func TestAliasLoadNull(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "aliases.json")}
	if err := os.WriteFile(store.Path, []byte("null"), 0o644); err != nil {
		t.Fatal(err)
	}
	names, records, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 || len(records) != 0 {
		t.Fatalf("names=%#v records=%#v", names, records)
	}
}

func TestAliasSaveErrors(t *testing.T) {
	store := Store{Path: t.TempDir()}
	if err := store.save(map[string]Record{}); err == nil {
		t.Fatal("expected write error")
	}
	parentFile := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(parentFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	store = Store{Path: filepath.Join(parentFile, "aliases.json")}
	if err := store.save(map[string]Record{}); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestAliasErrorBranches(t *testing.T) {
	dirPath := t.TempDir()
	store := Store{Path: dirPath}
	if err := store.Set("x", treehash.Hash("0123456789abcdef0123456789abcdef01234567")); err == nil {
		t.Fatal("expected set load error")
	}
	if _, _, err := store.Get("x"); err == nil {
		t.Fatal("expected get load error")
	}
	if _, _, err := store.List(); err == nil {
		t.Fatal("expected list load error")
	}

	badJSON := Store{Path: filepath.Join(t.TempDir(), "aliases.json")}
	if err := os.WriteFile(badJSON.Path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := badJSON.Get("x"); err == nil {
		t.Fatal("expected json error")
	}

	parentFile := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(parentFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	badSave := Store{Path: filepath.Join(parentFile, "aliases.json")}
	if err := badSave.Set("x", treehash.Hash("0123456789abcdef0123456789abcdef01234567")); err == nil {
		t.Fatal("expected save mkdir error")
	}
}
