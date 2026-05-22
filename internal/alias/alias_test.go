package alias

import (
	"path/filepath"
	"testing"

	"github.com/spmurray/gitsnap/internal/treehash"
)

func TestAliasSetGetList(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "aliases.json")}
	hash := treehash.Hash("0123456789abcdef0123456789abcdef01234567")

	if err := store.Set("session-1", hash); err != nil {
		t.Fatal(err)
	}
	got, ok, err := store.Get("session-1")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got != hash {
		t.Fatalf("Get = %q, %v; want %q, true", got, ok, hash)
	}

	names, records, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "session-1" {
		t.Fatalf("names = %#v", names)
	}
	if records["session-1"].Hash != hash {
		t.Fatalf("record hash = %q", records["session-1"].Hash)
	}
}
