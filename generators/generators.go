package generators

import (
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
)

const ManifestFilename string = ".accio.toml"

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

type Template interface {
	Body() []byte
	Filename() string
	Skip() bool
}

type Writer interface {
	WriteFile(name string, writer io.Writer) error
}

type Reader interface {
	ReadFile(name string) (io.Reader, error)
}

type Walker interface {
	Walk(root string, walkFn filepath.WalkFunc) error
}

type ReaderWalker interface {
	Reader
	Walker
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

func (g *Generator) ReadConfig(r io.Reader) error {
	byt, err := ioutil.ReadAll(r)
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

func (g *Generator) ManifestPath() string {
	return path.Join(g.Dest, ManifestFilename)
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
