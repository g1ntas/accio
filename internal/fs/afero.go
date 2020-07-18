package fs

import (
	"github.com/spf13/afero"
	"os"
)

// AferoFileTreeReader provides a primitive API to read directory and it's files,
// without caring about underlying filesystem and it's structure.
type AferoFileTreeReader struct {
	fs afero.Fs
}

// NewAferoFileTreeReader returns a new reader for specified path.
func NewAferoFileTreeReader(fs afero.Fs, base string) AferoFileTreeReader {
	return AferoFileTreeReader{fs: afero.NewBasePathFs(fs, base)}
}

// Walk walks the file tree, calling walkFn for each file or directory
// in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func (r AferoFileTreeReader) Walk(walkFn func(filepath string, isDir bool, err error) error) error {
	return afero.Walk(r.fs, "", func(path string, info os.FileInfo, err error) error {
		return walkFn(path, info.IsDir(), err)
	})
}

// ReadFile reads the file from file tree named by filename and returns
// the contents. A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from
// Read as an error to be reported.
func (r AferoFileTreeReader) ReadFile(filename string) ([]byte, error) {
	return afero.ReadFile(r.fs, filename)
}
