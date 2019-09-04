package generator

import (
	"os"
	"path/filepath"
)

type FileSystemRepository struct {
	Path       string
	Generators map[string]*Generator
}

func NewFileSystemRepository(path string) *FileSystemRepository {
	return &FileSystemRepository{Path: path}
}

func (r *FileSystemRepository) Dest() string {
	return r.Path
}

func (r *FileSystemRepository) Parse() error {
	err := filepath.Walk(r.Dest(), func(path string, info os.FileInfo, err error) error {
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
			r.addGenerator(gen)
		}
		// if path is a root directory, scan one level deeper
		if path == r.Dest() {
			return nil
		}
		return filepath.SkipDir
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *FileSystemRepository) addGenerator(g *Generator) {
	r.Generators[g.Name] = g
}
