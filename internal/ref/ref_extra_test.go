package ref

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spmurray/gitsnap/internal/alias"
)

func TestResolveAliasStoreError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(alias.Store{Path: path}, "x"); err == nil {
		t.Fatal("expected error")
	}
}
