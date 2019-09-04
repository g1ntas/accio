package generator

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Registry struct {
	Repos []*FileSystemRepository `json:"repositories"`
	path string
}

func NewRegistry(path string) *Registry {
	return &Registry{path: path}
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
		if err := os.MkdirAll(dir, os.ModeDir); err != nil {
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
func (r *Registry) AddRepository(repo *FileSystemRepository) {
	r.Repos = append(r.Repos, repo)
}

func (g *Generator) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Dest string `json:"dest"`
	} {
		Dest: g.Dest(),
	})
}
