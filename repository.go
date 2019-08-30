package main

import (
	"os"
	"path/filepath"
)

type fileSystemRepository struct {
	Origin string `json:"origin"`
	Dest string `json:"destination"`
	Generators map[string]*Generator `json:"generators"`
}

// repo := NewLocalRepo("~/code/symfony-crud")
// repo.Parse()
// registry.add(repo)
// registry.Save()

func NewLocalRepo(origin string) *fileSystemRepository {
	// todo: check origin is existing directory
	// todo: expand origin to absolute path and set as Dest
	return &fileSystemRepository{Origin: origin, Dest: origin}
}

func (r *fileSystemRepository) Parse() error {
	err := filepath.Walk("", processRepositoryPath)
	if err != nil {
		return err
	}
	// todo: check if `.accio.json` exist in root directory
	// 		todo: if so, parse config and add generator to repository
	// todo: if no, walk over directory looking for 1st-level subdirectories with `.accio.json` config file
		// todo: parse each config and save generator to repo
}

func processRepositoryPath(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	return nil
}