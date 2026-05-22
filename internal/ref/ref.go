package ref

import (
	"fmt"

	"github.com/spmurray/gitsnap/internal/alias"
	"github.com/spmurray/gitsnap/internal/treehash"
)

func Resolve(store alias.Store, raw string) (treehash.Hash, error) {
	if h, ok := treehash.Parse(raw); ok {
		return h, nil
	}
	h, ok, err := store.Get(raw)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("unknown snapshot or alias %q", raw)
	}
	return h, nil
}
