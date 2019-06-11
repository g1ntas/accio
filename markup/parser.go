package markup

import (
	"fmt"
	"runtime"
	"unicode"
)

// attribute names for reserved 'delimiters' tag
const (
	attrDelimitersLeft  = "left"
	attrDelimitersRight  = "right"
)

// TagNode todo
type TagNode struct {
	Attributes map[string]*AttrNode
	Body *string
	Name string
}

// AttrNode todo
type AttrNode struct {
	Tag *TagNode
	Name string
	Value string
}

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// Parser is the representation of a single parsed template.
type Parser struct {
	Name string     // name of the template represented by the tree.
	Tags []*TagNode // list of nodes of the tree.
	text string     // text parsed to create the template.
	// For parsing only; cleared after parse.
	lex *lexer
	token token  // token currently being parsed.
	tag *TagNode // tag currently being built.
	schema *Schema
}

// parse returns a parse.Parser of the template. If an error is encountered,
// parsing stops and an empty map is returned with error.
func Parse(name, text, leftDelim, rightDelim string) (p *Parser, err error) {
	p = &Parser{Name: name}
	p.text = text
	_, err = p.parse(text, leftDelim, rightDelim)
	return
}

// recover is the handler that turns panic into returns from the top level of parse.
func (p *Parser) recover(errp *error) {
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

// stop terminates Parser.
func (p *Parser) stop() {
	p.lex = nil
}

// parse parses the template definition string to construct a representation of
// the template for execution. If either body delimiter string is empty, the
// default ("<<" or ">>") is used.
func (p *Parser) parse(text, leftDelim, rightDelim string) (tree *Parser, err error) {
	defer p.recover(&err)
	p.Tags = []*TagNode{}
	p.lex = lex(p.Name, text, leftDelim, rightDelim) // start parsing
	p.text = text
	p.parseTemplate()
	p.stop()
	return p, nil
}

// parseTemplate is the top-level Parser for a template. It runs to EOF.
func (p *Parser) parseTemplate() {
	switch token := p.next(); token.typ {
	case tokenEOF:
		return
	case tokenDelimiters:
		p.parseDelimitersTag()
	case tokenIdentifier:
		p.parseTag()
	case tokenError:
		p.errorf("%s", token)
	default:
		p.errorf("unexpected %s", token)
	}
}

// parseTag todo
func (p *Parser) parseTag() {
	p.newTag(p.token.val)
	// consume next whitespace token
	switch token := p.next(); token.typ {
	case tokenEOF:
		return
	case tokenNewline:
		p.parseTemplate()
		return
	case tokenSpace:
		p.parseAttrOrBody()
		return
	default:
		p.errorf("unexpected %s", token)
	}
}

// parseDelimitersTag todo
func (p *Parser) parseDelimitersTag() {
	if len(p.Tags) > 0 {
		p.errorf("reserved tag %s is not allowed here, it must be defined before all other tags", p.token)
		return
	}
	for {
		switch token := p.next(); token.typ {
		case tokenAttrDeclare:
			p.parseDelimiterAttr()
			continue
		case tokenSpace:
			continue
		case tokenNewline:
		case tokenEOF:
			return
		case tokenLeftDelim:
			p.errorf("body is not allowed here", token)
		case tokenError:
			p.errorf("unexpected %s", token)
		}
	}
}

// parseDelimiterAttr todo
func (p *Parser) parseDelimiterAttr() {
	name, value := p.scanAttr()
	if (name == leftDelimiter || name == rightDelimiter) && containsInvisibleChars(value) {
		p.errorf("attribute %s of the tag %s can not contain invisible characters", name, p.tag.Name)
	}
	switch name {
	case attrDelimitersLeft:
		p.lex.leftDelim = value
		return
	case attrDelimitersRight:
		p.lex.rightDelim = value
		return
	}
}

// parseAttrOrBody todo
func (p *Parser) parseAttrOrBody() {
	for {
		switch token := p.next(); token.typ {
		case tokenSpace:
			continue
		case tokenEOF:
		case tokenNewline:
			return
		case tokenAttrDeclare:
			p.parseAttr()
		case tokenLeftDelim:
			p.parseBody()
			return
		default:
			p.errorf("unexpected %s", token)
		}
	}
}

// parseAttr todo
func (p *Parser) parseAttr() {
	name, value := p.scanAttr()
	if _, exists := p.tag.Attributes[name]; exists {
		p.errorf("attribute '%s' already exists for this tag", name)
	}
	// todo: perform schema validation
	p.tag.Attributes[name] = &AttrNode{
		Tag: p.tag,
		Name: name,
		Value: value,
	}
}

// scanAttr scans and consumes tokens of the attribute which is known to be present
// and returns it's name and value.
func (p *Parser) scanAttr() (name string, value string){
	token := p.next()
	if token.typ != tokenIdentifier {
		p.errorf("expected identifier, got %s", token)
	}
	name = token.val
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
	return
}

// parseBody todo
func (p *Parser) parseBody() {
	token := p.next()
	if token.typ == tokenNewline {
		token = p.next()
	}
	if token.typ != tokenBody {
		p.errorf("unexpected %s", token)
	}
	p.tag.Body = &token.val
	token = p.next()
	if token.typ != tokenRightDelim {
		p.errorf("unexpected %s", token)
	}
}

// next returns the next token.
func (p *Parser) next() token {
	p.token = p.lex.nextToken()
	return p.token
}

// errorf formats the error and terminates processing.
func (p *Parser) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("template at %s:%d: %s", p.Name, p.token.line, format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing with error.
func (p *Parser) error(err error) {
	p.errorf("%s", err)
}

// newTag todo
func (p *Parser) newTag(name string) *TagNode {
	p.tag = &TagNode{Name: name}
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

// containsInvisibleChars checks whether string contains any character
// which is not visible to the human eye, even if it consumes space at
// the screen.
func containsInvisibleChars(s string) bool {
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) {
			return true
		}
	}
	return false
}