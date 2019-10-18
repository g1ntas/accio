package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type NativeFS struct {
}

func NewNativeFS() *NativeFS {
	return &NativeFS{}
}

func (fs *NativeFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	// create dir if doesn't exist
	return ioutil.WriteFile(name, data, perm)
}

func (fs *NativeFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func (fs *NativeFS) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
