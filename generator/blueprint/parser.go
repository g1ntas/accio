package blueprint

import (
	"fmt"
	"github.com/cbroglie/mustache"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"strconv"
	"strings"

	"github.com/g1ntas/accio/markup"
)

const (
	tagFilename = "filename"
	tagSkip     = "skipif"
	tagTemplate = "template"
	tagPartial  = "partial"
	tagVariable = "variable"
	attrName    = "name"
)

// context carries data to be used in starlark scripts and mustache templates.
type context struct {
	vars     map[string]starlark.Value
	partials map[string]string
}

// newContext creates a context with provided data, and
// transforms that data to starlark compatible data types.
func newContext(data map[string]interface{}) (context, error) {
	ctx := context{
		vars:     make(map[string]starlark.Value),
		partials: make(map[string]string),
	}
	for k, v := range data {
		val, err := newValue(v)
		if err != nil {
			return context{}, err
		}
		ctx.vars[k] = val
	}
	return ctx, nil
}

// copy makes a new copy of context and all it's data,
// so changes to it, won't affect original context.
func (ctx context) copy() context {
	vars := make(map[string]starlark.Value)
	partials := make(map[string]string)
	for k, v := range ctx.vars {
		vars[k] = v
	}
	for k, v := range ctx.partials {
		partials[k] = v
	}
	ctx.vars = vars
	ctx.partials = partials
	return ctx
}

// varsDict returns context variables as starlark dictionary.
func (ctx *context) varsDict() (*starlark.Dict, error) {
	dict := starlark.NewDict(len(ctx.vars))
	for k, v := range ctx.vars {
		err := dict.SetKey(starlark.String(k), v)
		if err != nil {
			return nil, err
		}
	}
	return dict, nil
}

// varsGoMap returns context variables as golang map.
func (ctx *context) varsGoMap() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for k, v := range ctx.vars {
		goval, err := parseValue(v)
		if err != nil {
			return nil, err
		}
		m[k] = goval
	}
	return m, nil
}

type ParseError struct {
	Msg, Tag string
	Line     int
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("failed to parse tag %q on line %d: %s", err.Tag, err.Line, err.Msg)
}

func newErr(msg string, tag string, line int) *ParseError {
	return &ParseError{msg, tag, line}
}

// evalErr creates a new ParseError from external error.
func evalErr(tag *markup.TagNode, err error) error {
	if err == nil {
		return nil
	}
	newErr := &ParseError{Tag: tag.Name, Line: tag.Line}
	switch e := err.(type) {
	case syntax.Error:
		newErr.Msg = e.Msg
		newErr.Line = evalErrLine(tag, int(e.Pos.Line)-1)
	case resolve.Error:
		newErr.Msg = e.Msg
		newErr.Line = evalErrLine(tag, int(e.Pos.Line)-1)
	case resolve.ErrorList:
		return evalErr(tag, e[0])
	default:
		newErr.Msg = e.Error()
	}
	return newErr
}

type Parser struct {
	ctx context
	mp  *markup.Parser
}

func NewParser(d map[string]interface{}) (*Parser, error) {
	ctx, err := newContext(d)
	if err != nil {
		return nil, err
	}
	return &Parser{ctx: ctx}, nil
}

// blueprint is an alias for an anonymous struct used in
// generator.BlueprintParser interface.
type blueprint = struct {
	Body     string
	Filename string
	Skip     bool
}

func (p Parser) Parse(b []byte) (*blueprint, error) {
	var err error
	p.mp, err = markup.Parse(string(b), "", "")
	if err != nil {
		return nil, err
	}
	p.ctx = p.ctx.copy()
	return p.parse()
}

func (p *Parser) parse() (*blueprint, error) {
	var err error
	bp := &blueprint{}
	for _, tag := range p.mp.Tags {
		switch tag.Name {
		case tagVariable:
			err = p.parseVariable(tag)
			if err != nil {
				return nil, err
			}
		case tagFilename:
			bp.Filename, err = p.parseFilename(tag)
			if err != nil {
				return nil, err
			}
		case tagSkip:
			bp.Skip, err = p.parseSkip(tag)
			if err != nil {
				return nil, err
			}
		case tagPartial:
			err = p.parsePartial(tag)
			if err != nil {
				return nil, err
			}
		case tagTemplate:
			bp.Body, err = p.renderTemplate(tag)
			if err != nil {
				return nil, err
			}
		}
	}
	return bp, nil
}

func (p *Parser) parseVariable(tag *markup.TagNode) error {
	if hasEmptyBody(tag) {
		return nil
	}
	name := getAttr(tag, attrName)
	if isEmpty(name) {
		return nil
	}
	val, err := execute(parseScriptBody(tag), &p.ctx)
	if err != nil {
		return evalErr(tag, err)
	}
	p.ctx.vars[name] = val
	return nil
}

func (p *Parser) parseFilename(tag *markup.TagNode) (string, error) {
	if hasEmptyBody(tag) {
		return "", nil
	}
	v, err := execute(parseScriptBody(tag), &p.ctx)
	if err != nil {
		return "", evalErr(tag, err)
	}
	filename, err := parseString(v)
	if err != nil {
		return "", evalErr(tag, err)
	}
	return filename, nil
}

func (p *Parser) parseSkip(tag *markup.TagNode) (bool, error) {
	if hasEmptyBody(tag) {
		return false, nil
	}
	v, err := execute(parseScriptBody(tag), &p.ctx)
	if err != nil {
		return false, evalErr(tag, err)
	}
	return parseBool(v), nil
}

func (p *Parser) parsePartial(tag *markup.TagNode) error {
	if hasEmptyBody(tag) {
		return nil
	}
	name := getAttr(tag, attrName)
	if isEmpty(name) {
		return nil
	}
	p.ctx.partials[name] = tag.Body.Content
	return nil
}

func (p *Parser) renderTemplate(tag *markup.TagNode) (string, error) {
	data, err := p.ctx.varsGoMap()
	if err != nil {
		return "", err
	}
	provider := &mustache.StaticProvider{Partials: p.ctx.partials}
	var body string
	if tag.Body != nil {
		body = tag.Body.Content
	}
	content, err := mustache.RenderPartials(body, provider, data)
	if err != nil {
		msg, line := splitMustacheError(err)
		return "", newErr(msg, tag.Name, evalErrLine(tag, line))
	}
	return content, nil
}

func getAttr(tag *markup.TagNode, s string) string {
	for _, attr := range tag.Attributes {
		if attr.Name == s {
			return attr.Value
		}
	}
	return ""
}

// hasEmptyBody checks if tag contains empty body after trimming spaces.
func hasEmptyBody(t *markup.TagNode) bool {
	return t.Body == nil || isEmpty(t.Body.Content)
}

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// splitMustacheError takes error returned by mustache library and
// splits it into message and line on which error occurred.
// Mustache error is returned in format `line %d: %s` and is
// implemented in github.com/cbroglie/mustache/mustache.go:171.
// In case given error doesn't contain line in expected format,
// then original error message and line 0 is returned.
func splitMustacheError(err error) (string, int) {
	msg := err.Error()
	if msg[:5] != "line " {
		return msg, 0
	}
	i := strings.IndexRune(msg, ':')
	line, err := strconv.Atoi(msg[5:i])
	if err != nil {
		return msg, 0
	}
	msg = msg[i+2:] // remove colon and space following it
	return msg, line
}

func parseScriptBody(tag *markup.TagNode) string {
	if tag.Body == nil {
		return ""
	}
	if tag.Body.Inline {
		return wrapInlineScript(tag.Body.Content)
	}
	return tag.Body.Content
}

func evalErrLine(tag *markup.TagNode, line int) int {
	if tag.Body == nil || tag.Body.Inline {
		return tag.Line
	}
	return tag.Line + line
}
