package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	manifestFilename = ".accio.toml"
	templateExt      = ".accio"
)

type Generator struct {
	Dest    string
	Help    string    `toml:"help"`
	Prompts PromptMap `toml:"prompts"`
}

// OnExistsFn handles files that already exist at target path.
// Return true to overwrite file or false to skip it.
type OnExistsFn func(path string) bool

// OnErrorFn is called if error occurred when processing file.
// Return true to skip the file and continue process, or false
// to terminate Runner and return the error.
type OnErrorFn func(err error) bool

type blueprint = struct {
	Body     string
	Filename string
	Skip     bool
}

type BlueprintParser interface {
	Parse(b []byte) (*blueprint, error)
}

type Writer interface {
	WriteFile(name string, data []byte, perm os.FileMode) error
}

type Reader interface {
	ReadFile(name string) ([]byte, error)
}

type Walker interface {
	Walk(root string, walkFn filepath.WalkFunc) error
}

type ReaderWalker interface {
	Reader
	Walker
}

type Filesystem interface {
	Reader
	Walker
	Writer
	Stat(name string) (os.FileInfo, error)
}

// relFile represents file relative to generator's root directory.
type relFile struct {
	relpath string
	kind    os.FileMode
}

type RunError struct {
	Err error
	Path string
}

func (e *RunError) Error() string  {
	return fmt.Sprintf("generating %s: %s", e.Path, e.Err.Error())
}

func (e *RunError) Unwrap() error  {
	return e.Err
}

func IgnoreFile(p string) func(r *Runner) {
	return func(r *Runner) {
		r.ignore = append(r.ignore, &relFile{p, os.ModePerm})
	}
}

func IgnoreDir(p string) func(r *Runner) {
	return func(r *Runner) {
		r.ignore = append(r.ignore, &relFile{p, os.ModeDir})
	}
}

func OnFileExists(fn OnExistsFn) func(r *Runner) {
	return func(r *Runner) {
		r.onExists = fn
	}
}

func OnError(fn OnErrorFn) func(r *Runner) {
	return func(r *Runner) {
		r.onError = fn
	}
}

func NewGenerator(dir string) *Generator {
	return &Generator{
		Dest:    dir,
		Prompts: make(PromptMap),
	}
}

func (g *Generator) PromptAll(prompter Prompter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	// sort prompts by keys, so they always appear in the same order
	keys, i := make([]string, len(g.Prompts)), 0
	for k, _ := range g.Prompts {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		val, err := g.Prompts[k].Prompt(prompter)
		if err != nil {
			return map[string]interface{}{}, err
		}
		data[k] = val
	}
	return data, nil
}

func (g *Generator) manifestPath() string {
	return path.Join(g.Dest, manifestFilename)
}

type Runner struct {
	fs       Filesystem
	mp       BlueprintParser
	writeDir string // absolute path to the directory to write generated files
	onExists OnExistsFn
	onError  OnErrorFn
	ignore   []*relFile // collection of files to ignore during run
}

func NewRunner(fs Filesystem, mp BlueprintParser, dir string, options ...func(*Runner)) *Runner {
	r := &Runner{
		fs:       fs,
		mp:       mp,
		writeDir: dir,
		onExists: func(path string) bool {
			return false
		},
		onError: func(err error) bool {
			return false
		},
	}
	IgnoreFile(manifestFilename)(r)
	for _, option := range options {
		option(r)
	}
	return r
}

func (r *Runner) Run(generator *Generator) error {
	return r.fs.Walk(generator.Dest, func(abspath string, info os.FileInfo, err error) error {
		if err != nil {
			return r.handleError(err, abspath)
		}
		relpath, err := filepath.Rel(generator.Dest, abspath)
		if err != nil {
			return r.handleError(err, abspath)
		}
		// skip specified files and directories
		for _, f := range r.ignore {
			switch {
			case relpath != f.relpath:
				continue
			case info.Mode().IsDir() && f.kind.IsDir():
				return filepath.SkipDir
			case info.Mode().IsRegular() && f.kind.IsRegular():
				return nil
			}
		}
		// skip non-files
		if !info.Mode().IsRegular() {
			return nil
		}
		body, err := r.fs.ReadFile(abspath)
		if err != nil {
			return r.handleError(err, abspath)
		}
		target := filepath.Join(r.writeDir, relpath)
		if hasTemplateExtension(target) {
			target = target[:len(target)-len(templateExt)] // remove ext
			tpl, err := r.mp.Parse(body)
			switch {
			case err != nil:
				return r.handleError(err, abspath)
			case tpl.Skip:
				return nil
			case tpl.Filename != "":
				basename := filepath.Base(target)
				target = joinWithinRoot(r.writeDir, tpl.Filename)
				stat, err := r.fs.Stat(target)
				// if path is directory, then attach filename of source file
				if err == nil && stat.IsDir() {
					target = filepath.Join(target, basename)
				}
			}
			body = []byte(tpl.Body)
		}
		// if file exists, call callback to decide if it should be skipped
		if _, err := r.fs.Stat(target); err == nil && !r.onExists(target) {
			return nil
		}
		err = r.fs.WriteFile(target, body, 0775)
		if err != nil {
			return r.handleError(err, abspath)
		}
		return nil
	})
}

func (r *Runner) handleError(err error, path string) error {
	err = &RunError{err, path}
	if r.onError(err) {
		return nil
	}
	return err
}

func hasTemplateExtension(path string) bool {
	return len(path) > len(templateExt) && path[len(path)-len(templateExt):] == templateExt
}

// joinWithinRoot joins two paths ensuring that one (relative) path ends up
// inside the other (root) path. If relative path evaluates to be outside root
// directory, then it's treated as there's no parent directory and root is final.
func joinWithinRoot(root, relpath string) string {
	sep := string(filepath.Separator)
	parts := strings.Split(filepath.Clean(relpath), sep)
	for _, part := range parts {
		if part != ".." {
			break
		}
		parts = parts[1:]
	}
	return filepath.Join(root, strings.Join(parts, sep))
}
