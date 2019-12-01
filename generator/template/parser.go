package template // todo: rename to Model (ModelParser)

import (
	"github.com/cbroglie/mustache"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/markup"
)

type Parser struct {
	// Shared data to be used in scripts and templates.
	// It's not thread-safe, thus should not be modified blindly.
	ctx context
}

func NewParser(d map[string]interface{}) (*Parser, error) {
	ctx, err := newContext(d)
	if err != nil {
		return nil, err
	}
	return &Parser{ctx: ctx}, nil
}

func (p *Parser) Parse(b []byte) (*generator.Template, error) {
	parser, err := markup.Parse(string(b), "", "")
	if err != nil {
		return nil, err
	}
	m := parse(parser)

	ctx := p.copyContext()
	for _, variable := range m.Vars {
		v, err := execute(variable[1], ctx)
		if err != nil {
			return nil, err
		}
		ctx[variable[0]] = v
	}
	data, err := ctx.toGoMap()
	if err != nil {
		return nil, err
	}
	tpl := &generator.Template{}
	tpl.Body, err = renderTemplate(m, data)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func renderTemplate(m *Markup, data map[string]interface{}) (string, error) {
	provider := &mustache.StaticProvider{Partials: m.Partials}
	content, err := mustache.RenderPartials(m.Body, provider, data)
	if err != nil {
		return "", err
	}
	return content, nil
}

// copyContext makes a new copy of context, so it can be
// safely manipulated without side effects (e.g. race conditions)
func (p *Parser) copyContext() context {
	ctx := make(context)
	for k, v := range p.ctx {
		ctx[k] = v
	}
	return ctx
}
