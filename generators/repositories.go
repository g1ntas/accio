package generators

import (
	"errors"
	"io"
	"os"
	"path"
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

func (r *FileSystemRepository) Parse() (count int, err error) {
	err = filepath.Walk(r.Dest(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		gen := NewGenerator(path)
		err = gen.NewGeneratorFromConfig()
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

type GeneratorFinder struct {
	wlk Walker
	repo *FileSystemRepository
	paths []string
}

func (f *GeneratorFinder) FindManifests() ([]string, error) {
	err := f.wlk.Walk(f.repo.Dir(), f.findInRoot)
	if err != nil && err != io.EOF {
		return []string{}, err
	}
	return f.paths, nil
}

func (f *GeneratorFinder) findInRoot(pth string, mode os.FileMode, err error) error {
	if err != nil {
		return err
	}
	// manifests can only be located in root directory or one level deeper
	if mode.IsDir() && pth != f.repo.Dir() {
		return f.wlk.Walk(pth, f.findInSubdirectory)
	}
	if path.Base(pth) == ManifestFilename {
		f.paths = append(f.paths, pth)
		return io.EOF
	}
	return nil
}

func (f *GeneratorFinder) findInSubdirectory(pth string, mode os.FileMode, err error) error {
	if err != nil {
		return err
	}
	if mode.IsDir() {
		return filepath.SkipDir
	}
	if path.Base(pth) == ManifestFilename {
		f.paths = append(f.paths, pth)
		return io.EOF
	}
	return nil
}
