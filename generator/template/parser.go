package template // todo: rename to Model (ModelParser)

import (
	"github.com/cbroglie/mustache"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/markup"
	"strings"
)

type Parser struct {
	// Data to be used in scripts and templates. Shared between models.
	// It's not thread-safe, thus should not be modified directly.
	ctx context
}

type ParseError struct {
	Msg, Tag string
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
	err = parseVariables(m, ctx)
	if err != nil {
		return nil, err
	}
	tpl := &generator.Template{}
	tpl.Filename, err = parseFilename(m, ctx)
	if err != nil {
		return nil, err
	}
	tpl.Skip, err = parseSkip(m, ctx)
	if err != nil {
		return nil, err
	}
	tpl.Body, err = renderTemplate(m, ctx)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func parseVariables(m *schema, ctx context) error {
	const key = 0
	const value = 1
	for _, v := range m.vars {
		val, err := execute(v[value], ctx)
		if err != nil {
			return err
		}
		ctx[v[key]] = val
	}
	return nil
}

func parseFilename(m *schema, ctx context) (string, error) {
	if len(strings.TrimSpace(m.filename)) > 0 {
		v, err := execute(m.filename, ctx)
		if err != nil {
			return "", err
		}
		filename, err := parseString(v)
		if err != nil {
			return "", err
		}
		return filename, nil
	}
	return "", nil
}

func parseSkip(m *schema, ctx context) (bool, error) {
	if len(strings.TrimSpace(m.skip)) > 0 {
		v, err := execute(m.skip, ctx)
		if err != nil {
			return false, err
		}
		return parseBool(v), nil
	}
	return false, nil
}

func renderTemplate(m *schema, ctx context) (string, error) {
	data, err := ctx.toGoMap()
	if err != nil {
		return "", err
	}
	provider := &mustache.StaticProvider{Partials: m.partials}
	content, err := mustache.RenderPartials(m.body, provider, data) // todo: parse line from mustache.parseError
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
