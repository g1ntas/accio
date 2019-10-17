package generator

import (
	"github.com/BurntSushi/toml"
	"path"
	"path/filepath"
)

const manifestFilename string = ".accio.toml"

type Generator struct {
	Dest        string
	Name        string    `toml:"full-name"`
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

type Writer interface {
	WriteFile(name string, data []byte) error
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

type ReaderWalkerWriter interface {
	Reader
	Walker
	Writer
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
