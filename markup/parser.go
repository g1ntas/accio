package markup

import (
	"fmt"
	"runtime"
)

// Tag todo
type Tag struct {
	Attributes map[string]string
	Body string
	Name string
}

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// parser is the representation of a single parsed template.
type parser struct {
	Name string // name of the template represented by the tree.
	Tags []*Tag // list of nodes of the tree.
	text string // text parsed to create the template.
	// For parsing only; cleared after parse.
	lex *lexer
	token token // token currently being parsed.
	tag *Tag // tag currently being builded.
}

// Parse returns a parse.parser of the template. If an error is encountered,
// parsing stops and an empty map is returned with error.
func Parse(name, text, leftDelim, rightDelim string) (*parser, error) {
	t := &parser{Name: name}
	t.text = text
	_, err := t.Parse(text, leftDelim, rightDelim)
	return t, err
}

// recover is the handler that turns panic into returns from the top level of Parse.
func (p *parser) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if p != nil {
			p.lex.drain()
			p.stop()
		}
		*errp = e.(error)
	}
}

// stop terminates parser.
func (p *parser) stop() {
	p.lex = nil
}

// Parse parses the template definition string to construct a representation of
// the template for execution. If either body delimiter string is empty, the
// default ("<<" or ">>") is used.
func (p *parser) Parse(text, leftDelim, rightDelim string) (tree *parser, err error) {
	defer p.recover(&err)
	p.Tags = []*Tag{}
	p.lex = lex(p.Name, text, leftDelim, rightDelim) // start parsing
	p.text = text
	p.parseTemplate()
	p.stop()
	return p, nil
}

// parseTemplate is the top-level parser for a template. It runs to EOF.
func (p *parser) parseTemplate() {
	for {
		switch token := p.next(); token.typ {
		case tokenEOF:
			return
		case tokenIdentifier:
			p.parseTag()
			continue
		case tokenError:
			return // todo: error

		}

	}
}

// parseTag todo
func (p *parser) parseTag() {
	// todo: check reserved words; In this case 'delimiters': must be at the beginning of template.
	p.newTag(p.token.val)
	// consume next whitespace token
	switch token := p.next(); token.typ {
	case tokenEOF:
	case tokenNewline:
		return
	case tokenSpace:
		p.parseAttrOrBody()
	default:
		p.errorf("unexpected %s", token)
	}
}

// parseAttrOrBody todo
func (p *parser) parseAttrOrBody() {
	switch token := p.next(); token.typ {
	case tokenEOF:
	case tokenNewline:
		return
	case tokenAttrDeclare:
		p.parseAttr()
	case tokenLeftDelim:
		p.parseBody()
	default:
		p.errorf("unexpected %s", token)
	}
}

// parseAttr todo
func (p *parser) parseAttr() {
	token := p.next()
	if token.typ != tokenIdentifier {
		p.errorf("expected identifier, got %s", token)
	}
	name := token.val
	if _, exists := p.tag.Attributes[name]; exists {
		p.errorf("attribute '%s' already exists for this tag", name)
	}
	if token = p.next(); token.typ != tokenAssign {
		p.errorf("expected '=', got %s", token)
	}
	token = p.next()
	if token.typ != tokenString {
		p.errorf("expected quoted string, got %s", token)
	}
	value, err := unquoteString(token.val)
	if err != nil {
		p.error(err)
	}
	// todo: validate if value is valid based on tag name and attr name
	p.tag.Attributes[name] = value
}

// parseBody todo
func (p *parser) parseBody() {
	// todo
}

// next returns the next token.
func (p *parser) next() token {
	return p.lex.nextToken()
}

// errorf formats the error and terminates processing.
func (p *parser) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("template at %s:%d: %s", p.Name, p.token.line, format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing with error.
func (p *parser) error(err error) {
	p.errorf("%s", err)
}

// newTag todo
func (p *parser) newTag(name string) *Tag {
	p.tag = &Tag{Name: name}
	p.Tags = append(p.Tags, p.tag)
	return p.tag
}

// unquoteString removes double quote characters surrounding the string.
func unquoteString(s string) (string, error) {
	l := len(s)
	if s[0:1] != "\"" || s[l-1:] != "\"" {
		return "", fmt.Errorf("value is expected to be surrounded with \" quotes")
	}
	return s[1:l], nil
}