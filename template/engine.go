package template

type Template struct {
}

func (t *Template) Body() []byte {
	return []byte{}
}

func (t *Template) Filename() string {
	return ""
}

func (t *Template) Skip() bool {
	return false
}

type Engine struct{
}

func (e *Engine) Parse(b []byte, data map[string]interface{}) (Template, error) {
	return &Template{}, nil
}

