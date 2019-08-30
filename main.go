package main

import (
	"encoding/json"
	"github.com/g1ntas/accio/cmd"
	"io/ioutil"
	"os"
	"path/filepath"
)

type registry struct {
	Repos []*fileSystemRepository `json:"repositories"`
	path string
}

var Registry *registry

func main() {
	var err error
	path, err := registryPath()
	if err != nil {
		handleError(err)
		return
	}
	Registry, err = loadRegistry(path)
	if err != nil {
		handleError(err)
		return
	}
	cmd.Execute()
}

func handleError(err error) {
	panic(err)
}

func newRegistry(path string) *registry {
	return &registry{path: path}
}

// loadRegistry reads registry file from user config directory.
// If registry file does not exist within user config directory,
// then new registry struct is returned instead.
func loadRegistry(path string) (*registry, error) {
	reg := newRegistry(path)
	b, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return reg, nil
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, reg)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// Save stores all the data from registry struct into configured registry file.
func (r *registry) Save() error {
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
func (r *registry) addRepository(repo *fileSystemRepository) {
	r.Repos = append(r.Repos, repo)
}

// userConfigDir returns base directory for storing user configuration for accio.
func userConfigDir() (string, error) {
	// todo: in go 1.13 os.UserConfigDir() should be added, when released use it as a base dir
	// todo: see: https://github.com/golang/go/issues/29960
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".accio"), nil
}

// registryPath returns file path to the config file containing information about existing generators and repositories.
func registryPath() (string, error) {
	dir, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "registry.json"), nil
}
