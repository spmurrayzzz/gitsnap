package ref

import (
	"path/filepath"
	"testing"

	"github.com/spmurray/gitsnap/internal/alias"
	"github.com/spmurray/gitsnap/internal/treehash"
)

func TestResolveHash(t *testing.T) {
	hash := "0123456789abcdef0123456789abcdef01234567"
	got, err := Resolve(alias.Store{Path: filepath.Join(t.TempDir(), "aliases.json")}, hash)
	if err != nil {
		t.Fatal(err)
	}
	if got != treehash.Hash(hash) {
		t.Fatalf("got %q", got)
	}
}

func TestResolveAlias(t *testing.T) {
	store := alias.Store{Path: filepath.Join(t.TempDir(), "aliases.json")}
	hash := treehash.Hash("0123456789abcdef0123456789abcdef01234567")
	if err := store.Set("turn-1", hash); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve(store, "turn-1")
	if err != nil {
		t.Fatal(err)
	}
	if got != hash {
		t.Fatalf("got %q", got)
	}
}

func TestResolveUnknown(t *testing.T) {
	_, err := Resolve(
		alias.Store{Path: filepath.Join(t.TempDir(), "aliases.json")},
		"missing",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
