package markup

import "runtime"

// Tree is the representation of a single parsed template.
type Tree struct {
	Name string // name of the template represented by the tree.
	ParseName string // name of the top-level template during parsing, for error messages.
	Tags *[]Tag // list of tag nodes of the tree.
	text string // text parsed to create the template.
	// Parsing only; cleared after parse.
	lex *lexer
	token token // token currently being parsed.
}

// Parse returns a parse.Tree of the template. If an error is encountered,
// parsing stops and an empty map is returned with error.
func Parse(name, text, leftDelim, rightDelim string) (*Tree, error) {
	t := New(name)
	t.text = text
	_, err := t.Parse(text, leftDelim, rightDelim)
	return t, err
}

// New allocates a new parse tree with given name.
func New(name string) *Tree {
	return &Tree{Name: name}
}

// recover is the handler that turns panic into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.lex.drain()
			t.stopParse()
		}
		*errp = e.(error)
	}
}

// stopParse terminates parsing.
func (t *Tree) stopParse() {
	t.lex = nil
}

// Parse parses the template definition string to construct a representation of
// the template for execution. If either body delimiter string is empty, the
// default ("<<" or ">>") is used.
func (t *Tree) Parse(text, leftDelim, rightDelim string) (tree *Tree, err error) {
	defer t.recover(&err)
	t.ParseName = t.Name
	t.Tags = &[]Tag{}
	t.lex = lex(t.Name, text, leftDelim, rightDelim) // start parsing
	t.text = text
	t.parse()
	t.stopParse()
	return t, nil
}

// parse is the top-level parser for a template. It runs to EOF.
func (t *Tree) parse() {
	for {
		switch t.next().typ {
		case tokenEOF:
			return;
		case tokenIdentifier:
			// parse tag
		
		}

	}
}

// next returns the next token.
func (t *Tree) next() token {
	return t.lex.nextToken()
}