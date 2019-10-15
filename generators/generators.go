package generators

import (
	"io/ioutil"
	"os"
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
	WriteFile(filename string, content []byte) error
}

type Reader interface {
	ReadFile(path string) ([]byte, error)
}

type WalkFunc func(path string, info os.FileMode, err error) error
type Walker interface {
	Walk(root string, walkFn WalkFunc) error
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

/*func NewGeneratorFromConfig(r io.Reader) error {
	err = toml.Unmarshal(r, &g)
	if err != nil {
		return err
	}
	// todo: validate parsed data
	return nil
}
*/
func NewGeneratorFromConfig(dest string) error {
	info, err := os.Stat(dest)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return &os.PathError{
			Op:   "parse manifest",
			Path: dest,
			Err:  errors.New("is a directory, but expected a file"),
		}
	}
	byt, err := ioutil.ReadFile(dest)
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
