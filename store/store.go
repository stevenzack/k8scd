package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Store struct {
	dir string
}

func NewStore(dir string) (*Store, error) {
	e := os.MkdirAll(dir, 0700)
	if e != nil {
		return nil, e
	}

	return &Store{
		dir: dir,
	}, nil
}
func (s *Store) getValue(key string) ([]byte, error) {
	b, e := os.ReadFile(filepath.Join(s.dir, key))
	if e != nil && os.IsNotExist(e) {
		return nil, nil
	}
	return b, nil
}
func (s *Store) setValue(key string, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	dst := filepath.Join(s.dir, key)
	e = os.MkdirAll(filepath.Dir(dst), 0700)
	if e != nil {
		return e
	}

	fo, e := os.OpenFile(dst, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if e != nil {
		return e
	}
	defer fo.Close()
	_, e = fo.Write(b)
	return e
}
