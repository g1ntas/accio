package template // todo: rename to Model (ModelParser)

import (
	"errors"
	"fmt"
	"github.com/cbroglie/mustache"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/markup"
	"go.starlark.net/starlark"
	"strconv"
	"strings"
)

const (
	tagFilename = "filename"
	tagSkip     = "skipif"
	tagTemplate = "template"
	tagPartial  = "partial"
	tagVariable = "variable"
	attrName    = "name"
)

// tagsPriority defines order in which tags should be parsed,
// lower number means higher priority.
var tagsPriority = map[string]uint{
	tagVariable: 0,
	tagFilename: 1,
	tagSkip:     1,
	tagPartial:  1,
	tagTemplate: 2,
}

// context carries data to be used in starlark scripts and mustache templates.
type context struct {
	vars map[string]starlark.Value
	partials map[string]string
}

// newContext creates new context with predefined starlark data.
func newContext(data map[string]interface{}) (context, error) {
	ctx := context{
		vars: make(map[string]starlark.Value),
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

// varsToDict returns variables as starlark dictionary.
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

// varsToGoMap returns variables as go map.
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

type Parser struct {
	ctx context
}

type ParseError struct {
	Op, Msg, Tag string
	Line         int
}

func (err ParseError) Error() string {
	// Msg: unmatched open tag
	// Tag: template
	// Line: 1
	// failed to parse tag 'template' because error occurred on line 1 (unmatched open tag)
	/*msg := err.Msg
	if len(err.Tag) != 0 {
		msg = fmt.Sprintf("%s", err.Msg)
	}
	if err.Line != 0 {

	}*/
	return fmt.Sprintf("%s", err.Msg)
}

func newErr(msg string) ParseError {
	return ParseError{Msg: msg}
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
	ctx := p.ctx.copy()
	return parse(parser, &ctx)
}

func parse(p *markup.Parser, ctx *context) (*generator.Template, error) {
	var err error
	tpl := &generator.Template{}
	for _, tag := range orderTags(p.Tags) {
		switch tag.Name {
		case tagVariable:
			err = parseVariable(tag, ctx)
			if err != nil {
				return nil, err
			}
		case tagFilename:
			tpl.Filename, err = parseFilename(tag, ctx)
			if err != nil {
				return nil, err
			}
		case tagSkip:
			tpl.Skip, err = parseSkip(tag, ctx)
			if err != nil {
				return nil, err
			}
		case tagPartial:
			err = parsePartial(tag, ctx)
			if err != nil {
				return nil, err
			}
		case tagTemplate:
			tpl.Body, err = renderTemplate(tag, ctx)
			if err != nil {
				return nil, err
			}
		}
	}
	return tpl, nil
}

// orderTags sorts tag nodes in their parse order.
// Implements bubble sort algorithm.
func orderTags(tags []*markup.TagNode) []*markup.TagNode {
	n := len(tags)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if tagsPriority[tags[j].Name] > tagsPriority[tags[j+1].Name] {
				tags[j], tags[j+1] = tags[j+1], tags[j]
			}
		}
	}
	return tags
}

func parseVariable(tag *markup.TagNode, ctx *context) error {
	if hasEmptyBody(tag) {
		return nil
	}
	name := getAttr(tag, attrName)
	if isEmpty(name) {
		return nil
	}
	val, err := execute(parseScriptBody(tag), ctx)
	if err != nil {
		return err
	}
	ctx.vars[name] = val
	return nil
}

func parsePartial(tag *markup.TagNode, ctx *context) error {
	if hasEmptyBody(tag) {
		return nil
	}
	name := getAttr(tag, attrName)
	if isEmpty(name) {
		return nil
	}
	ctx.partials[name] = tag.Body.Content
	return nil
}

func parseFilename(tag *markup.TagNode, ctx *context) (string, error) {
	if hasEmptyBody(tag) {
		return "", nil
	}
	v, err := execute(parseScriptBody(tag), ctx)
	if err != nil {
		return "", err
	}
	filename, err := parseString(v)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func parseSkip(tag *markup.TagNode, ctx *context) (bool, error) {
	if hasEmptyBody(tag) {
		return false, nil
	}
	v, err := execute(parseScriptBody(tag), ctx)
	if err != nil {
		return false, err
	}
	return parseBool(v), nil
}

func renderTemplate(tag *markup.TagNode, ctx *context) (string, error) {
	data, err := ctx.varsGoMap()
	if err != nil {
		return "", err
	}
	provider := &mustache.StaticProvider{Partials: ctx.partials}
	var body string
	if tag.Body != nil {
		body = tag.Body.Content
	}
	content, err := mustache.RenderPartials(body, provider, data)
	if err != nil {
		msg, _ := splitMustacheError(err)
		return "", errors.New(msg)
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
// then original error message and line number 0 is returned.
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
	msg = msg[i+2:] // remove colon, and space going after it
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