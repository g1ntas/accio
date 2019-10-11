package generators

import (
	"errors"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"path/filepath"
)

const ManifestFilename string = ".accio.toml"

var templateEngine TemplateEngine

type Generator struct {
	Dest        string
	Name        string    `toml:"full-name"`
	Description string    `toml:"description"`
	Help        string    `toml:"help"`
	Prompts     promptMap `toml:"prompts"`
}

type GeneratorError struct {
	Op   string
	Name string
	Path string
	Err  error
}

type TemplateEngine interface {
	Parse(b []byte, data map[string]interface{}) (Template, error)
}

type Template interface {
	Body() []byte
	Filename() string
	Skip() bool
}

type Writer interface {
	Write(filename string, content []byte) error
}

func (e *GeneratorError) Error() string {
	return e.Op + " " + e.Name + " at " + e.Path + ": " + e.Err.Error()
}

func NewGenerator(dir string) *Generator {
	return &Generator{Dest: dir}
}

func (g *Generator) wrapErr(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &GeneratorError{operation, g.Name, g.Dest, err}

}

func (g *Generator) ParseManifest() error {
	path := filepath.Join(g.Dest, ManifestFilename)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return &os.PathError{
			Op:   "parse manifest",
			Path: path,
			Err:  errors.New("is a directory, but expected a file"),
		}
	}
	byt, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(byt, &g)
	if err != nil {
		return err
	}
	// todo: validate parsed data
	return nil
}

func (g *Generator) Run(prompter Prompter, writer Writer) error {
	manifestPath := filepath.Join(g.Dest, ManifestFilename)
	data, err := g.prompt(prompter)
	if err != nil {
		return err
	}
	return filepath.Walk(g.Dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || path == manifestPath {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		filename, err := filepath.Rel(g.Dest, path)
		if err != nil {
			return err
		}
		if filepath.Ext(info.Name()) == ".accio" {
			tpl, err := templateEngine.Parse(b, data)
			if err != nil {
				return err
			}
			if f := tpl.Filename(); f != "" {
				filename = f
			}
			if tpl.Skip() {
				return nil
			}
			err = writer.Write(filename, tpl.Body())
			if err != nil {
				return err
			}
			return nil
		}
		err = writer.Write(filename, b)
		if err != nil {
			return err
		}
		return nil
	})
}

func (g *Generator) prompt(prompter Prompter) (data map[string]interface{}, err error) {
	for key, p := range g.Prompts {
		val, err := p.prompt(prompter)
		if err != nil {
			return map[string]interface{}{}, err
		}
		data[key] = val
	}
	return
}
