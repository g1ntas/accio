package main

import (
	"errors"
	"os"
	"path/filepath"
)

type fileSystemRepository struct {
	Origin string `json:"origin"`
	Dest string `json:"destination"`
	Generators map[string]*Generator `json:"generators"`
}

var errCfgIsDir = errors.New("config path is a directory")

// repo := NewLocalRepo("~/code/symfony-crud")
// repo.Parse()
// registry.add(repo)
// registry.Save()

func NewFileSystemRepository(origin string) *fileSystemRepository {
	// todo: check origin is existing directory
	// todo: expand origin to absolute path and set as Dest
	return &fileSystemRepository{Origin: origin, Dest: origin}
}

func (r *fileSystemRepository) Parse() error {
	err := filepath.Walk(r.Dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		gen, err := parseDir(path)
		// todo: should it be more strict and instead forbid missing config?
		if err != nil && err != errCfgIsDir && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			r.addGenerator(gen)
		}
		// if path is a root directory, scan one level deeper
		if path == r.Dest{
			return nil
		}
		return filepath.SkipDir
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *fileSystemRepository) addGenerator(g *Generator) {
	r.Generators[g.Name] = g
}

func parseDir(path string) (*Generator, error) {
	cfgPath := filepath.Join(path, ConfigFilename)
	f, err := os.Stat(cfgPath)
	if err != nil {
		return nil, err
	}
	if f.IsDir() {
		return nil, errCfgIsDir
	}
	return parseConfig(cfgPath)
}