package generators

import (
	"os"
	"path"
	"path/filepath"
)

type Filesystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte) error
	Walk(root string, walkFn filepath.WalkFunc) error
}

type TemplateEngine interface {
	Parse(b []byte, data map[string]interface{}) (Template, error)
	Extension() string
}

type Runner struct {
	prompter Prompter
	fs Filesystem
	tplEngine TemplateEngine
}

func NewRunner(p Prompter, fs Filesystem, tpl TemplateEngine) *Runner {
	return &Runner{p, fs, tpl}
}

func (r *Runner) Run(gen *Generator, forceOverwrite bool) error {
	manifestPath := path.Join(gen.Dest, ManifestFilename)
	data, err := gen.prompt(r.prompter)
	if err != nil {
		return err
	}
	return r.fs.Walk(gen.Dest, func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || pth == manifestPath {
			return nil
		}
		b, err := r.fs.ReadFile(pth)
		if err != nil {
			return err
		}
		filename, err := filepath.Rel(gen.Dest, pth)
		if err != nil {
			return err
		}
		// todo: pass cwd
		if r.hasTemplateExtension(pth) {
			tpl, err := r.tplEngine.Parse(b, data)
			if err != nil {
				return err
			}
			if f := tpl.Filename(); f != "" {
				filename = f
			}
			if tpl.Skip() {
				return nil
			}
			err = r.fs.WriteFile(filename, tpl.Body())
			if err != nil {
				return err
			}
			return nil
		}
		err = r.fs.WriteFile(filename, b)
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *Runner) hasTemplateExtension(path string) bool {
	ext := "." + r.tplEngine.Extension()
	l := len(path)
	return path[l-len(ext):l] == ext
}