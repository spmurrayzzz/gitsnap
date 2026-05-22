package alias

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spmurray/gitsnap/internal/treehash"
)

type Record struct {
	Hash      treehash.Hash `json:"hash"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Store struct {
	Path string
}

func (s Store) Set(name string, hash treehash.Hash) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[name] = Record{Hash: hash, UpdatedAt: time.Now().UTC()}
	return s.save(m)
}

func (s Store) Get(name string) (treehash.Hash, bool, error) {
	m, err := s.load()
	if err != nil {
		return "", false, err
	}
	rec, ok := m[name]
	return rec.Hash, ok, nil
}

func (s Store) List() ([]string, map[string]Record, error) {
	m, err := s.load()
	if err != nil {
		return nil, nil, err
	}
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, m, nil
}

func (s Store) load() (map[string]Record, error) {
	b, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	var m map[string]Record
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]Record{}
	}
	return m, nil
}

func (s Store) save(m map[string]Record) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	b = append(b, '\n')
	return os.WriteFile(s.Path, b, 0o644)
}
