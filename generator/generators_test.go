package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// file info mock implementation
type fileMock struct {
	path    string
	mode    os.FileMode
	content string
}

var _ os.FileInfo = (*fileMock)(nil)

func (f *fileMock) Name() string       { return filepath.Base(f.path) }
func (f *fileMock) Mode() os.FileMode  { return f.mode }
func (f *fileMock) IsDir() bool        { return f.mode.IsDir() }
func (f *fileMock) Size() int64        { return 0 }
func (f *fileMock) ModTime() time.Time { return time.Time{} }
func (f *fileMock) Sys() interface{}   { return nil }
func (f *fileMock) String() string {
	if f == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{path: %q, content: %q}", f.path, f.content)
}

func file(filename string, content string) *fileMock {
	return &fileMock{filepath.FromSlash(filename), 0775, content}
}
func symlink(filename string) *fileMock {
	return &fileMock{filepath.FromSlash(filename), os.ModeSymlink, ""}
}
func dir(dirname string) *fileMock {
	return &fileMock{filepath.FromSlash(dirname), os.ModeDir, ""}
}

// tree is meant to simulate files structure within filesystem
type tree []*fileMock

func (t tree) get(path string) *fileMock {
	for _, f := range t {
		if f.path == path {
			return f
		}
	}
	return nil
}

// filesystem mock implementation
type fsMock struct {
	files tree
	output tree
}

var _ Filesystem = (*fsMock)(nil)

func (fs *fsMock) ReadFile(filename string) ([]byte, error) {
	if f := fs.files.get(filename); f != nil {
		return []byte(f.content), nil
	}
	return []byte{}, os.ErrNotExist
}

func (fs *fsMock) WriteFile(name string, data []byte, _ os.FileMode) error {
	f := file(name, string(data))
	fs.output = append(fs.output, f)
	fs.files = append(fs.files, f)
	return nil
}

func (fs *fsMock) Walk(root string, walkFn filepath.WalkFunc) error {
	for _, f := range fs.files {
		if len(f.path) < len(root) || f.path[:len(root)] != root {
			continue
		}
		if err := walkFn(f.path, f, nil); err != nil && err != filepath.SkipDir {
			return err
		}
	}
	return nil
}

func (fs *fsMock) Stat(name string) (os.FileInfo, error) {
	if f := fs.files.get(name); f != nil {
		return f, nil
	}
	return nil, os.ErrNotExist
}

// template engine mock implementation
type tplEngineMock struct{}

var _ TemplateEngine = (*tplEngineMock)(nil)

func (e *tplEngineMock) Parse(b []byte, _ map[string]interface{}) (*Template, error) {
	tpl := &Template{}
	err := json.Unmarshal(b, tpl)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func equalTrees(t1, t2 tree) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i, f := range t1 {
		if f == nil || t2[i] == nil {
			return false
		}
		if f.path != t2[i].path {
			return false
		}
		if f.content != t2[i].content {
			return false
		}
		if f.mode != t2[i].mode {
			return false
		}
	}
	return true
}

var runnerTests = []struct {
	name   string
	input  tree
	output tree
}{
	{"no files", tree{}, tree{}},
	{"ignore directories", tree{dir("a")}, tree{}},
	{"ignore symbolic links", tree{symlink("a")}, tree{}},
	{"ignore manifest", tree{file("generator/.accio.toml", "")}, tree{}},
	{"write static file",
		tree{file("generator/a.txt", "test")},
		tree{file("output/a.txt", "test")}},
	{"write multiple files",
		tree{file("generator/a.txt", "file1"), file("generator/b.txt", "file2")},
		tree{file("output/a.txt", "file1"), file("output/b.txt", "file2")}},
	{"write nested files",
		tree{dir("generator/abc"), file("generator/abc/test.txt", "file")},
		tree{file("output/abc/test.txt", "file")}},
	{"write template file",
		tree{file("generator/test.txt.accio", `{"body": "test"}`)},
		tree{file("output/test.txt", "test")}},
	{"template | skip file",
		tree{file("generator/test.txt.accio", `{"skip": true}`)},
		tree{}},
	{"template | custom filename",
		tree{file("generator/test.txt.accio", `{"filename": "custom.txt", "body": "---"}`)},
		tree{file("output/custom.txt", "---")}},
	{"template | nested custom filename",
		tree{file("generator/test.txt.accio", `{"filename": "dir/custom.txt"}`)},
		tree{file("output/dir/custom.txt", "")}},
	{"template | append static name if filename is directory",
		tree{dir("output/abc"), file("generator/test.txt.accio", `{"filename": "abc"}`)},
		tree{file("output/abc/test.txt", "")}},
	{"template | don't write outside root",
		tree{file("generator/test.txt.accio", `{"filename": "../../../custom.txt"}`)},
		tree{file("output/custom.txt", "")}},
}

func TestRunner(t *testing.T) {
	gen := &Generator{Dest: "generator"}
	for _, test := range runnerTests {
		fs := &fsMock{test.input, tree{}}
		runner := NewRunner(nil, fs, &tplEngineMock{})
		err := runner.Run(gen, "output", false)
		switch {
		case err != nil:
			t.Errorf("%s:\nunexpected error: %v", test.name, err)
		case !equalTrees(test.output, fs.output):
			t.Errorf("%s:\ngot:\n\t%v\nexpected:\n\t%v", test.name, fs.output, test.output)
		}
	}
}

// todo: handle errors
// todo: handle existing files (using external functions)
