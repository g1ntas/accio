package generators

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
)

const SerializationFormat = "gob"

type Registry struct {
	Repos map[string]*FileSystemRepository `json:"repositories"`
}

func NewRegistry() *Registry {
	return &Registry{
		Repos: make(map[string]*FileSystemRepository),
	}
}

func ReadRegistry(r io.Reader) (*Registry, error) {
	reg := NewRegistry()
	dec := gob.NewDecoder(r)
	err := dec.Decode(reg)
	if err != nil {
		return nil, err
	}
	return reg, err
}

func WriteRegistry(w io.Writer, reg *Registry) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(reg)
	if err != nil {
		return err
	}
	return nil
}

// addRepository adds a repository to the registry.
func (r *Registry) AddRepository(repo *FileSystemRepository) error {
	if _, ok := r.Repos[repo.Path]; ok {
		// todo: make own error OriginError?
		return &os.PathError{Op: "add repository", Path: repo.Path, Err: errors.New("already exists")}
	}
	r.Repos[repo.Path] = repo
	return nil
}
