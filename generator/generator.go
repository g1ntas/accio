package generator

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	manifestFilename = ".accio.toml"
	templateExt      = ".accio"
)

type Generator struct {
	Dest        string
	Name        string    `toml:"name"`
	Description string    `toml:"description"`
	Help        string    `toml:"help"`
	Prompts     PromptMap `toml:"prompts"`
}

type GeneratorError struct {
	Op   string
	Name string
	Path string
	Err  error
}

type model = struct {
	Body     string
	Filename string
	Skip     bool
}

type ModelParser interface {
	Parse(b []byte) (*model, error)
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

func (e *GeneratorError) Error() string {
	return e.Op + " " + e.Name + " at " + e.Path + ": " + e.Err.Error()
}

func NewGenerator(dir string) *Generator {
	return &Generator{
		Dest:    dir,
		Prompts: make(PromptMap),
	}
}

func (g *Generator) PromptAll(prompter Prompter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for key, p := range g.Prompts {
		val, err := p.Prompt(prompter)
		if err != nil {
			return map[string]interface{}{}, err
		}
		data[key] = val
	}
	return data, nil
}

func (g *Generator) wrapErr(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &GeneratorError{operation, g.Name, g.Dest, err}

}

func (g *Generator) manifestPath() string {
	return path.Join(g.Dest, manifestFilename)
}

type Runner struct {
	fs       Filesystem
	mp       ModelParser
	writeDir string // absolute path to the directory to write generated files

	// Confirmation callback to overwrite existing files.
	// Returning true overwrites files, and false skips them.
	overwriteFunc func(path string) bool
}

// todo: implement functional options
func NewRunner(fs Filesystem, mp ModelParser, dir string, overwriteFunc func(path string) bool) *Runner {
	return &Runner{
		fs:            fs,
		mp:            mp,
		writeDir:      dir,
		overwriteFunc: overwriteFunc,
	}
}

func (r *Runner) Run(generator *Generator) error {
	return r.fs.Walk(generator.Dest, func(abspath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// ignore non-files and manifest
		if !info.Mode().IsRegular() || abspath == generator.manifestPath() {
			return nil
		}
		body, err := r.fs.ReadFile(abspath)
		if err != nil {
			return err
		}
		relpath, err := filepath.Rel(generator.Dest, abspath)
		if err != nil {
			return err
		}
		target := filepath.Join(r.writeDir, relpath)
		if hasTemplateExtension(target) {
			target = target[:len(target)-len(templateExt)] // remove ext
			tpl, err := r.mp.Parse(body)
			switch {
			case err != nil:
				return err
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
		if _, err := r.fs.Stat(target); err == nil && !r.overwriteFunc(target) {
			return nil
		}
		err = r.fs.WriteFile(target, body, 0775)
		if err != nil {
			return err
		}
		return nil
	})
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
