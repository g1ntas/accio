package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Filesystem interface {
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Walk(root string, walkFn filepath.WalkFunc) error
	Stat(name string) (os.FileInfo, error)
}

type NativeFS struct {
}

func NewNativeFS() *NativeFS {
	return &NativeFS{}
}

func (fs *NativeFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(name), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, data, perm)
}

func (fs *NativeFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func (fs *NativeFS) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (fs *NativeFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func NewDryFS(fs Filesystem) *DryFS {
	return &DryFS{fs}
}

type DryFS struct {
	Filesystem
}

func (fs *DryFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return nil
}