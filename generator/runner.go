package generator

import (
	"os"
	"path"
	"strings"
)

const templateExt = ".accio"

type Template interface {
	Body() []byte
	Filename() string
	Skip() bool
}

type TemplateEngine interface {
	Parse(b []byte, data map[string]interface{}) (Template, error)
}

type Runner struct {
	prompter Prompter
	fs ReaderWalkerWriter
	tplEngine TemplateEngine
}

func NewRunner(p Prompter, fs ReaderWalkerWriter, tpl TemplateEngine) *Runner {
	return &Runner{p, fs, tpl}
}

func (r *Runner) Run(gen *Generator, writeDir string, forceOverwrite bool) error {
	manifestPath := path.Join(gen.Dest, manifestFilename)
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
		filename := strings.Replace(pth, gen.Dest, "", 1)
		if err != nil {
			return err
		}
		if r.hasTemplateExtension(pth) {
			tpl, err := r.tplEngine.Parse(b, data)
			if err != nil {
				return err
			}
			if f := tpl.Filename(); f != "" {
				filename = path.Join(gen.Dest, f)
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
	l := len(path)
	return path[l-len(templateExt):l] == templateExt
}