package generators

import (
	"errors"
	"os"
	"path/filepath"
)

type Repository struct {
	Path       string
	Generators map[string]*Generator
}

func NewRepository(path string) *Repository {
	return &Repository{
		Path:       path,
		Generators: make(map[string]*Generator),
	}
}

func (r *Repository) Dir() string {
	return r.Path
}

func (r *Repository) ImportGenerators(wr ReaderWalker) (count int, err error) {
	gen, err := parseGeneratorDir(wr, r.Dir())
	// if root directory doesnt contain generator, then check subdirectories
	if os.IsNotExist(err) {
		err = r.importGeneratorsFromSubdirs(wr)
		if err != nil {
			return 0, err
		}
		return len(r.Generators), nil
	}
	if err != nil {
		return 0, err
	}
	if err = r.addGenerator(gen); err != nil {
		return 0, err
	}
	return 1, nil
}

func (r *Repository) importGeneratorsFromSubdirs(wr ReaderWalker) error {
	return wr.Walk(r.Dir(), func(dest string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || dest == r.Dir() {
			return nil
		}
		gen, err := parseGeneratorDir(wr, dest)
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.SkipDir
			}
			return err
		}
		if err := r.addGenerator(gen); err != nil {
			return err
		}
		return filepath.SkipDir
	})
}

func parseGeneratorDir(r Reader, dir string) (*Generator, error) {
	gen := NewGenerator(dir)
	b, err := r.ReadFile(gen.ManifestPath())
	if err != nil {
		return nil, err
	}
	err = gen.ReadConfig(b)
	if err != nil {
		return nil, err
	}
	return gen, nil
}

func (r *Repository) addGenerator(g *Generator) error {
	if _, ok := r.Generators[g.Name]; ok {
		return g.wrapErr("add generator", errors.New("already exists within same repository"))
	}
	r.Generators[g.Name] = g
	return nil
}