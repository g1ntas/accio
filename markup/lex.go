// inspired by https://golang.org/src/text/template/parse/lex.go

package markup

import (
	"fmt"
	"unicode/utf8"
)

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType
	pos Pos
	value string
	line int
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.value
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.value)
	case len(i.value) > 10:
		return fmt.Sprintf("%.10q...", i.value)
	}
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF // end of file
	itemWS // whitespace
	itemSpace // single space
	itemIdentifier // identifier for tags and attributes

	itemBool // boolean literal
	itemNumber // number literal (including imaginary)
	itemString // string literal
	itemBody // raw tag body text between left and right delimiters

	itemAttribute // dash ('-') introducing an attribute declaration
	itemAssign // equals sign ('=') introducing an attribute assignment

	// Keywords
	itemKeyword // used only to delimit the keywords
	itemTagMain // main tag keyword
	itemAttrLeftDelimiter // left delimiter attribute keyword
	itemAttrRightDelimiter // right delimiter attribute keyword
)

var keywords = map[string]itemType{
	"aml": itemTagMain,
	"open": itemAttrLeftDelimiter,
	"end": itemAttrRightDelimiter,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name string // the name of the input; used only for error reports
	input string // the string being scanned
	leftDelimiter string // start of tag
	rightDelimiter string // end of tag
	position Pos // current position in the input
	start Pos // start position of this item
	width Pos // width of last rune read from input
	items chan item // channel of scanned items
	line int // 1+number of newlines seen
	startLine int // start line of this item
}

// next returns the next rune in the input.
func (lex *lexer) next() rune {
	if int(lex.position) >= len(lex.input) {
		lex.width = 0
		return eof
	}
	char, width := utf8.DecodeRuneInString(lex.input[lex.position:])
	lex.width = Pos(width)
	lex.position += lex.width
	if char == '\n' {
		lex.line++
	}
	return char
}

// peek returns but does not consume the next rune in the input.
func (lex *lexer) peek() rune {
	char := lex.next()
	// todo: implement peek separately
	lex.backup()
	return char
}

// backup steps back one rune. Can only be called once per call of next.
func (lex *lexer) backup() {
	lex.position -= lex.width
	if lex.width == 1 && lex.input[lex.position] == '\n' {
		// todo: doesn't reset lex.width
		lex.line--
	}
}

// emit passes an item back to the client.
func (lex *lexer) emit(typ itemType) {
	lex.items <- item{typ, lex.start, lex.input[lex.start:lex.position], lex.startLine}
	lex.start = lex.position
	lex.startLine = lex.line
}