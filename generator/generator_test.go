package generator

import (
	"encoding/json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// fsOpFn performs operation on the given filesystem.
type fsOpFn func(afero.Fs) error

// assertFn performs test assertion on the given filesystem.
type assertFn func(*testing.T, afero.Fs)

// blueprintBlueprintMock represents BlueprintParser implementation.
type blueprintParserMock struct{}

var noOutput = []assertFn{doesntExist("/output")}

var _ BlueprintParser = (*blueprintParserMock)(nil)

// Parse decodes json into Blueprint
func (e *blueprintParserMock) Parse(b []byte) (*blueprint, error) {
	tpl := &blueprint{}
	err := json.Unmarshal(b, tpl)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

const (
	skipExisting      = true
	overwriteExisting = false
)

var runnerTests = []struct {
	name         string
	input        []fsOpFn   // sequence of FS operations to prepare initial generator directory structure
	output       []assertFn // sequence of test assertions to perform after runner completes
	skipExisting bool
}{
	{
		"no files",
		[]fsOpFn{},
		noOutput,
		skipExisting,
	},
	{
		"ignore directories",
		[]fsOpFn{dir("/generator/a")},
		noOutput,
		skipExisting,
	},
	{
		"ignore symbolic links",
		[]fsOpFn{symlink("/generator/a")},
		noOutput,
		skipExisting,
	},
	{
		"ignore manifest",
		[]fsOpFn{file("/generator/.accio.toml", "")},
		noOutput,
		skipExisting,
	},
	{
		"write static file",
		[]fsOpFn{file("/generator/a.txt", "test")},
		[]assertFn{fileExists("/output/a.txt", "test")},
		skipExisting,
	},
	{
		"write multiple files",
		[]fsOpFn{file("/generator/a.txt", "file1"), file("/generator/b.txt", "file2")},
		[]assertFn{fileExists("/output/a.txt", "file1"), fileExists("/output/b.txt", "file2")},
		skipExisting,
	},
	{
		"write nested files",
		[]fsOpFn{dir("/generator/abc"), file("/generator/abc/test.txt", "file")},
		[]assertFn{fileExists("/output/abc/test.txt", "file")},
		skipExisting,
	},
	{
		"write blueprint file",
		[]fsOpFn{file("/generator/test.txt.accio", `{"body": "test"}`)},
		[]assertFn{fileExists("/output/test.txt", "test")},
		skipExisting,
	},
	{
		"blueprint | skip file",
		[]fsOpFn{file("/generator/test.txt.accio", `{"skip": true}`)},
		noOutput,
		skipExisting,
	},
	{
		"blueprint | custom filename",
		[]fsOpFn{file("/generator/test.txt.accio", `{"filename": "custom.txt", "body": "---"}`)},
		[]assertFn{fileExists("/output/custom.txt", "---")},
		skipExisting,
	},
	{
		"blueprint | nested custom filename",
		[]fsOpFn{file("/generator/test.txt.accio", `{"filename": "dir/custom.txt"}`)},
		[]assertFn{fileExists("/output/dir/custom.txt", "")},
		skipExisting,
	},
	{
		"blueprint | append static name if filename is directory",
		[]fsOpFn{dir("/output/abc"), file("/generator/test.txt.accio", `{"filename": "abc"}`)},
		[]assertFn{fileExists("/output/abc/test.txt", "")},
		skipExisting,
	},
	{
		"blueprint | don't write outside root",
		[]fsOpFn{file("/generator/test.txt.accio", `{"filename": "../../../custom.txt"}`)},
		[]assertFn{fileExists("/output/custom.txt", "")},
		skipExisting,
	},
	{
		"blueprint | write to root",
		[]fsOpFn{file("/generator/abc/test.txt.accio", `{"filename": "custom.txt"}`)},
		[]assertFn{fileExists("/output/custom.txt", "")},
		skipExisting,
	},
	{
		"overwrite if file exists",
		[]fsOpFn{file("/generator/test.txt", "new"), file("/output/test.txt", "old")},
		[]assertFn{fileExists("/output/test.txt", "new")},
		overwriteExisting,
	},
	{
		"skip if file exists",
		[]fsOpFn{file("/generator/test.txt", "new"), file("/output/test.txt", "old")},
		[]assertFn{fileExists("/output/test.txt", "old")},
		skipExisting,
	},
}

// file creates a file with a specified content at target path.
func file(filename string, content string) fsOpFn {
	return func(fs afero.Fs) error {
		return afero.WriteFile(fs, filename, []byte(content), 0775)
	}
}

// dir creates a directory at target path.
func dir(dirname string) fsOpFn {
	return func(fs afero.Fs) error {
		return fs.Mkdir(dirname, 0755)
	}
}

// symlink creates a symlink file at target path.
func symlink(filename string) fsOpFn {
	return func(fs afero.Fs) error {
		return afero.WriteFile(fs, filename, []byte{}, os.ModeSymlink)
	}
}

// doesntExist asserts that target path doesn't exist.
func doesntExist(filename string) assertFn {
	return func(t *testing.T, fs afero.Fs) {
		exists, err := afero.Exists(fs, filename)
		require.NoError(t, err)
		require.Falsef(t, exists, "file %s exists", filename)
	}
}

// fileExists asserts if target file exists and contains given content.
func fileExists(filename string, content string) assertFn {
	return func(t *testing.T, fs afero.Fs) {
		b, err := afero.ReadFile(fs, filename)
		if os.IsNotExist(err) {
			require.FailNowf(t, "expected file or directory", "file %s doesn't exist", filename)
		}
		require.NoError(t, err)
		require.Equal(t, content, string(b))
	}
}

func TestRunner(t *testing.T) {
	gen := &Generator{Dest: "/generator"}
	for _, test := range runnerTests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			err := fs.Mkdir("/generator", 0755)
			require.NoError(t, err)

			// build initial FS structure
			for _, fsOperation := range test.input {
				err := fsOperation(fs)
				assert.NoError(t, err)
			}
			runner := NewRunner(fs, &blueprintParserMock{}, "/output")
			runner.onExists = func(p string) bool {
				return !test.skipExisting
			}
			err = runner.Run(gen)
			require.NoError(t, err)

			// assert final FS structure
			for _, assertion := range test.output {
				assertion(t, fs)
			}
		})
	}
}
