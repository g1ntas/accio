package generator

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

const ManifestFilename string = ".accio.json"

type Generator struct {
	dest        string
	Name        string
	Description string
	Prompts     []interface{}
}

func (g *Generator) Dest() string {
	return g.dest
}

func NewGenerator(dir string) *Generator {
	return &Generator{dest: dir}
}

func (g *Generator) ParseManifest() error {
	path := filepath.Join(g.Dest(), ManifestFilename)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return &os.PathError{
			Op:   "open",
			Path: path,
			Err:  errors.New("is a directory, but expected a file"),
		}
	}
	byt, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byt, g)
	if err != nil {
		return err
	}
	// todo: validate parsed data
	return nil
}