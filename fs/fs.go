package fs

import (
	"io/ioutil"
	"path/filepath"
)

type NativeFS struct {

}

func NewNativeFS() *NativeFS {
	return &NativeFS{}
}

func (fs *NativeFS) WriteFile(name string, data []byte) error {
	return ioutil.WriteFile(name, data, 0775)
}

func (fs *NativeFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func (fs *NativeFS) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
