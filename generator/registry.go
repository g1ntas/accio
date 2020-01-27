package generator

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
)

type Registry struct {
	Repos map[string]*Repository
}

func init() {
	// register Prompt interface types
	gob.Register(&input{})
	gob.Register(&integer{})
	gob.Register(&confirm{})
	gob.Register(&list{})
	gob.Register(&choice{})
	gob.Register(&multiChoice{})
}

func NewRegistry() *Registry {
	return &Registry{
		Repos: make(map[string]*Repository),
	}
}

// addRepository adds a repository to the registry.
func (r *Registry) AddRepository(repo *Repository) error {
	if _, ok := r.Repos[repo.Path]; ok {
		// todo: make own error OriginError?
		return &os.PathError{Op: "add repository", Path: repo.Path, Err: errors.New("already exists")}
	}
	r.Repos[repo.Path] = repo
	return nil
}

func (r *Registry) FindGenerators(name string) []*Generator {
	matches := make([]*Generator, 0, 1)
	for _, repo := range r.Repos {
		if g, ok := repo.Generators[name]; ok {
			matches = append(matches, g)
		}
	}
	return matches
}

func Deserialize(r io.Reader) (*Registry, error) {
	reg := NewRegistry()
	dec := gob.NewDecoder(r)
	err := dec.Decode(reg)
	if err != nil {
		return nil, err
	}
	return reg, err
}

func Serialize(w io.Writer, reg *Registry) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(reg)
	if err != nil {
		return err
	}
	return nil
}