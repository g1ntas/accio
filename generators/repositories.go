package generators

import (
	"errors"
	"os"
	"path/filepath"
)

type FileSystemRepository struct {
	Path       string                `json:"path"`
	Generators map[string]*Generator `json:"generators"`
}

func NewFileSystemRepository(path string) *FileSystemRepository {
	return &FileSystemRepository{
		Path:       path,
		Generators: make(map[string]*Generator),
	}
}

func (r *FileSystemRepository) Dest() string {
	return r.Path
}

func (r *FileSystemRepository) Parse() (count int, err error) {
	err = filepath.Walk(r.Dest(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		gen := NewGenerator(path)
		err = gen.ParseManifest()
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			if err = r.addGenerator(gen); err != nil {
				return err
			}
			count++
		}
		// if path is a root directory, scan one level deeper
		if path == r.Dest() {
			return nil
		}
		return filepath.SkipDir
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *FileSystemRepository) addGenerator(g *Generator) error {
	if _, ok := r.Generators[g.Name]; ok {
		return g.wrapErr("add generator", errors.New("already exists within same repository"))
	}
	r.Generators[g.Name] = g
	return nil
}
