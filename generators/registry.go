package generators

import (
	"errors"
	"os"
)

type Registry struct {
	Repos map[string]*FileSystemRepository
}

func NewRegistry() *Registry {
	return &Registry{
		Repos: make(map[string]*FileSystemRepository),
	}
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
