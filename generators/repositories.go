package generators

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type FileSystemRepository struct {
	Path       string
	Generators map[string]*Generator
}

func NewFileSystemRepository(path string) *FileSystemRepository {
	return &FileSystemRepository{
		Path:       path,
		Generators: make(map[string]*Generator),
	}
}

func (r *FileSystemRepository) Dir() string {
	return r.Path
}

func (r *FileSystemRepository) ImportGenerators(wr ReaderWalker) (count int, err error) {
	gen, err := parseGeneratorDir(wr, r.Dir())
	if err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	if os.IsNotExist(err) {
		generators, err := parseGeneratorSubdirectories(wr, r.Dir())
		if err != nil {
			return 0, err
		}
	}
	os.IsNotExist(err)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func parseGeneratorSubdirectories(wr ReaderWalker, root string) (generators []*Generator, err error) {
	generators = make([]*Generator, 5, 0)
	err = wr.Walk(root, func(dest string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || dest == root {
			return nil
		}
		gen, err := parseGeneratorDir(wr, dest)
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.SkipDir
			}
			return err
		}
		generators = append(generators, gen)
		return filepath.SkipDir
	})
	return
}

func parseGeneratorDir(r Reader, dir string) (*Generator, error) {
	gen := NewGenerator(dir)
	reader, err := r.ReadFile(gen.ManifestPath())
	if err != nil {
		return nil, err
	}
	err = gen.ReadConfig(reader)
	if err != nil {
		return nil, err
	}
	return gen, nil
}

func (r *FileSystemRepository) addGenerator(g *Generator) error {
	if _, ok := r.Generators[g.Name]; ok {
		return g.wrapErr("add generator", errors.New("already exists within same repository"))
	}
	r.Generators[g.Name] = g
	return nil
}