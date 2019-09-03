package main

import (
	"encoding/json"
	"github.com/g1ntas/accio/cmd"
	"io/ioutil"
	"os"
	"path/filepath"
)

type registry struct {
	Repos []*FileSystemRepository `json:"repositories"`
	path string
}

var Registry *registry

func main() {
	var err error
	path, err := registryPath()
	if err != nil {
		handleError(err)
		return
	}
	Registry, err = loadRegistry(path)
	if err != nil {
		handleError(err)
		return
	}
	cmd.Execute()
}

func handleError(err error) {
	panic(err)
}