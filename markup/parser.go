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
	Body *Body
	Name string
	Line int
}

type Body struct {
	Content string
	Inline bool
}

// AttrNode todo
type AttrNode struct {
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
	Tags []*TagNode // list of nodes of the tree.
	text string     // text parsed to create the template.
	// For parsing only; cleared after parse.
	lex *lexer
	token token  // token currently being parsed.
	tag *TagNode // tag currently being built.
}

// parseStateFn represents the state of the parser as a function that returns the next state.
type parseStateFn func(*Parser) parseStateFn

// Parse is same as Parser.Parse, but also creates and returns Parser.
func Parse(text, leftDelim, rightDelim string) (p *Parser, err error) {
	p = NewParser()
	err = p.Parse(text, leftDelim, rightDelim)
	return
}

// NewParser constructs new parser.
func NewParser() *Parser {
	p := &Parser{}
	return p
}

// Parse parses given text string into list of nodes in Parser.Tags.
// If an error is encountered parsing stops and error is returned.
func (p *Parser) Parse(text, leftDelim, rightDelim string) (err error) {
	defer p.recover(&err)
	p.text = text
	p.startParse(lex(p.text, leftDelim, rightDelim))
	p.parse()
	p.stopParse()
	return nil
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
			p.stopParse()
		}
		*errp = e.(error)
	}
}

func (p *Parser) startParse(lex *lexer) {
	p.Tags = []*TagNode{}
	p.lex = lex
}

func (p *Parser) stopParse() {
	p.lex = nil
}

func (p *Parser) parse() {
	for state := parseTemplate; state != nil; {
		state = state(p)
	}
}

// parseTemplate is the top-level Parser for a template.
// It runs until it reaches EOF or error is encountered.
func parseTemplate(p *Parser) parseStateFn {
	switch token := p.next(); token.typ {
	case tokenEOF:
		return nil
	case tokenDelimiters:
		return parseDelimitersTag
	case tokenIdentifier:
		return parseTag
	case tokenError:
		return p.errorf("%s", token)
	default:
		return p.unexpected()
	}
}

func parseTag(p *Parser) parseStateFn {
	p.tag = &TagNode{Name: p.token.val, Line: p.token.line}
	p.tag.Attributes = make(map[string]*AttrNode)
	return parseAttrOrBody
}

func finishParsingTag(p *Parser) parseStateFn {
	p.Tags = append(p.Tags, p.tag)
	p.tag = nil
	switch p.token.typ {
	case tokenEOF:
		return nil
	case tokenNewline:
		return parseTemplate
	default:
		return p.unexpected()
	}
}

// parseAttrOrBody todo
func parseAttrOrBody(p *Parser) parseStateFn {
	switch token := p.next(); token.typ {
	case tokenEOF, tokenNewline:
		return finishParsingTag
	case tokenAttrDeclare:
		return parseAttr
	case tokenLeftDelim:
		return parseBody
	default:
		return p.unexpected()
	}
}

// parseDelimitersTag todo
func parseDelimitersTag(p *Parser) parseStateFn {
	if len(p.Tags) > 0 {
		return p.errorf("reserved tag %s is not allowed here, it must be defined before all other tags", p.token)
	}
	return parseDelimiterAttrs
}

// parseDelimiterAttr todo
func parseDelimiterAttrs(p *Parser) parseStateFn {
	attrs := make(map[string]string, 2)
	for p.next().typ == tokenAttrDeclare {
		name, value := p.scanAttr()
		if name != attrDelimitersLeft && name != attrDelimitersRight {
			return p.errorf("unexpected attribute '%s'", name)
		}
		if _, ok := attrs[name]; ok {
			return p.errorf("attribute '%s' is already defined", name)
		}
		if containsInvisibleChars(value) {
			return p.errorf("attribute '%s' must not contain any invisible character")
		}
		attrs[name] = value
	}
	var ok bool
	if p.lex.leftDelim, ok = attrs[attrDelimitersLeft]; !ok {
		return p.errorf("missing attribute '%s'", attrDelimitersLeft)
	}
	if p.lex.rightDelim, ok = attrs[attrDelimitersRight]; !ok {
		return p.errorf("missing attribute '%s'", attrDelimitersRight)
	}
	return parseTemplate
}

// parseAttr todo
func parseAttr(p *Parser) parseStateFn {
	name, value := p.scanAttr()
	if _, exists := p.tag.Attributes[name]; exists {
		return p.errorf("attribute '%s' already exists for this tag", name)
	}
	p.tag.Attributes[name] = &AttrNode{
		Name: name,
		Value: value,
	}
	return parseAttrOrBody
}

// scanAttr scans and consumes tokens of the attribute which is known to be present
// and returns it's full-name and value.
func (p *Parser) scanAttr() (string, string) {
	token := p.next()
	if token.typ != tokenIdentifier {
		p.errorf("expected identifier, got %s", token)
	}
	name := token.val
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
	return name, value
}

// parseBody todo
func parseBody(p *Parser) parseStateFn {
	p.tag.Body = &Body{}
	switch token := p.next(); token.typ {
	case tokenInlineBody:
		p.tag.Body.Inline = true
		fallthrough
	case tokenMultilineBody:
		p.tag.Body.Content = token.val
	default:
		return p.unexpected()
	}
	if token := p.next(); token.typ != tokenRightDelim {
		return p.unexpected()
	}
	switch token := p.next(); token.typ {
	case tokenEOF, tokenNewline:
		return finishParsingTag
	default:
		return p.unexpected()
	}
}

// next returns the next token.
func (p *Parser) next() token {
	p.token = p.lex.nextToken()
	return p.token
}

// errorf formats the error and terminates processing.
func (p *Parser) errorf(format string, args ...interface{}) parseStateFn {
	format = fmt.Sprintf("%s at line %d", format, p.token.line)
	panic(fmt.Errorf(format, args...))
	return nil
}

// error terminates processing with error.
func (p *Parser) error(err error) parseStateFn {
	return p.errorf("%s", err)
}

// error terminates processing with error.
func (p *Parser) unexpected() parseStateFn {
	return p.errorf("unexpected %s", p.token)
}

// unquoteString removes double quote characters surrounding the string.
func unquoteString(s string) (string, error) {
	l := len(s)
	if s[0:1] != "\"" || s[l-1:] != "\"" {
		return "", fmt.Errorf("value is expected to be surrounded with quotes")
	}
	return s[1:l-1], nil
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