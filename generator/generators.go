package generator

import (
	"github.com/BurntSushi/toml"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const manifestFilename string = ".accio.toml"
const templateExt = ".accio"

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

type Template struct {
	Body string
	Filename string
	Skip bool
}

type TemplateEngine interface {
	Parse(b []byte, data map[string]interface{}) (*Template, error)
}

type Runner struct {
	prompter  Prompter
	fs        Filesystem
	tplEngine TemplateEngine
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
		Dest: dir,
		Prompts: make(PromptMap),
	}
}

func (g *Generator) wrapErr(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &GeneratorError{operation, g.Name, g.Dest, err}

}

func (g *Generator) readConfig(b []byte) error {
	err := toml.Unmarshal(b, &g)
	if err != nil {
		return err
	}
	// todo: validate parsed data
	return nil
}

func (g *Generator) manifestPath() string {
	return path.Join(g.Dest, manifestFilename)
}

func (g *Generator) prompt(prompter Prompter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for key, p := range g.Prompts {
		val, err := p.prompt(prompter)
		if err != nil {
			return map[string]interface{}{}, err
		}
		data[key] = val
	}
	return data, nil
}

func NewRunner(p Prompter, fs Filesystem, tpl TemplateEngine) *Runner {
	return &Runner{p, fs, tpl}
}

func (r *Runner) Run(gen *Generator, writeDir string, forceOverwrite bool) error {
	manifestPath := path.Join(gen.Dest, manifestFilename)
	data, err := gen.prompt(r.prompter)
	if err != nil {
		return err
	}
	return r.fs.Walk(gen.Dest, func(abspath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// ignore non-files and manifest
		if !info.Mode().IsRegular() || abspath == manifestPath {
			return nil
		}
		b, err := r.fs.ReadFile(abspath)
		if err != nil {
			return err
		}
		relpath, err := filepath.Rel(gen.Dest, abspath)
		basename := filepath.Base(relpath)
		if err != nil {
			return err
		}
		if r.hasTemplateExtension(basename) {
			tpl, err := r.tplEngine.Parse(b, data)
			if err != nil {
				return err
			}
			basename = basename[:len(basename)-len(templateExt)]
			if f := tpl.Filename; f != "" {
				relpath = f
			} else {
				// remove template extension
				relpath = relpath[:len(relpath)-len(templateExt)]
			}
			if tpl.Skip {
				return nil
			}
			b = []byte(tpl.Body)
		}
		target := joinWithinRoot(writeDir, relpath)
		stat, err := r.fs.Stat(target)
		if err == nil && stat.IsDir() {
			target = filepath.Join(target, filepath.Base(basename))
		}
		err = r.fs.WriteFile(target, b, info.Mode())
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *Runner) hasTemplateExtension(path string) bool {
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
