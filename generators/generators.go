package generators

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

const ManifestFilename string = ".accio.json"

type Generator struct {
	Dest string `json:"dest"`
	Name string `json:"name"`
}

type GeneratorError struct {
	Op   string
	Name string
	Path string
	Err  error
}

func (e *GeneratorError) Error() string {
	return e.Op + " " + e.Name + " at " + e.Path + ": " + e.Err.Error()
}

func NewGenerator(dir string) *Generator {
	return &Generator{Dest: dir}
}

func (g *Generator) wrapErr(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &GeneratorError{operation, g.Name, g.Dest, err}

}

func (g *Generator) ParseManifest() error {
	// todo: use yaml instead
	path := filepath.Join(g.Dest, ManifestFilename)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return &os.PathError{
			Op:   "parse manifest",
			Path: path,
			Err:  errors.New("is a directory, but expected a file"),
		}
	}
	byt, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byt, &g)
	if err != nil {
		return err
	}
	// todo: validate parsed data
	return nil
}
