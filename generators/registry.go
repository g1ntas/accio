package generators

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Registry struct {
	Repos map[string]*FileSystemRepository `json:"repositories"`
	path  string
}

func NewRegistry(path string) *Registry {
	return &Registry{
		Repos: make(map[string]*FileSystemRepository),
		path: path,
	}
}

// Load reads registry file from user config directory and stores data in struct.
// If registry file does not exist within user config directory,
// then new registry struct is returned instead.
func (r *Registry) Load() error {
	b, err := ioutil.ReadFile(r.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, r)
	if err != nil {
		return err
	}
	return err
}

// Save stores all the data from registry struct into configured registry file.
func (r *Registry) Save() error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	err = ioutil.WriteFile(r.path, b, 0644)
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
