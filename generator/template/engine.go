package template

import "github.com/g1ntas/accio/generator"

type Engine struct{
}

func (e *Engine) Parse(b []byte, data map[string]interface{}) (*generator.Template, error) {
	tpl := &generator.Template{"", "", false}
	return tpl, nil
}

