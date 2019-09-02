package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

const ConfigFilename string = ".accio.json"

type Generator struct {
	Name string `json:"name"`
	Dest string `json:"destination"`
	Description string `json:"description"`
	Prompts []interface{}
}

func newGenerator(dir string) *Generator {
	return &Generator{Dest: dir}
}

func parseConfig(path string) (*Generator, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	gen := newGenerator(filepath.Dir(path))
	err = json.Unmarshal(b, gen)
	if err != nil {
		return nil, err
	}
	// todo: validate parsed data
	return gen, nil
}