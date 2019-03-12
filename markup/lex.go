// inspired by https://golang.org/src/text/template/parse/lex.go

package markup

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// tokenType represents the type of the Token
type tokenType int

const (
	tokenError      tokenType = iota // error occurred; value is text of error
	tokenEOF                         // end of file
	tokenWhitespace                  // whitespace
	tokenSpace                       // single space
	tokenIdentity                    // identity for tags and attributes

	tokenBool    // boolean literal
	tokenNumber  // number literal (including imaginary)
	tokenString  // string literal
	tokenBody    // raw tag body text between left and right delimiters

	tokenAttribute  // dash ('-') introducing an attribute declaration
	tokenAssign     // equals sign ('=') introducing an attribute assignment

	// Keywords
	tokenKeyword             // used only to delimit the keywords
	tokenTagMain             // main tag keyword
	tokenAttrLeftDelimiter   // left delimiter attribute keyword
	tokenAttrRightDelimiter  // right delimiter attribute keyword
)

var keywords = map[string]tokenType{
	"aml":   tokenTagMain,
	"start": tokenAttrLeftDelimiter,
	"end":   tokenAttrRightDelimiter,
}

const eof = -1

// todo: write doc
type token struct {
	typ  tokenType // the type of this token.
	pos  Pos       // the starting position, in bytes,  of this token in the input string.
	val  string    // the value of this token.
	line int       // the line number at the start of this token
}

func (t token) String() string {
	switch {
	case t.typ == tokenEOF:
		return "EOF"
	case t.typ == tokenError:
		return t.val
	case t.typ > tokenKeyword:
		return fmt.Sprintf("<%s>", t.val)
	case len(t.val) > 10:
		return fmt.Sprintf("%.10q...", t.val)
	}
	return fmt.Sprintf("%q", t.val)
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name      string     // the name of the input; used only for errors
	input     string     // the string being scanned
	pos       Pos        // current position in the input
	start     Pos        // start position of current token
	width     Pos        // width of last rune read from the input
	line      int        // 1+number of newlines seen
	startLine int        // start line of current token
	tokens    chan token // channel of scanned tokens
}

// atEndOfFile checks whether there are any characters left to scan.
func (lx *lexer) atEndOfFile() bool {
	return int(lx.pos) >= len(lx.input)
}

// next returns and does consume the next rune in the input.
func (lx *lexer) next() rune {
	if lx.atEndOfFile() {
		lx.width = 0
		return eof
	}
	rn, size := utf8.DecodeRuneInString(lx.input[lx.pos:])
	lx.width = Pos(size)
	lx.pos += lx.width
	if rn == '\n' {
		lx.line++
	}
	return rn
}

// peek returns but does not consume the next rune in the input.
func (lx *lexer) peek() rune {
	if lx.atEndOfFile() {
		return eof
	}
	r, _ := utf8.DecodeRuneInString(lx.input[lx.pos:])
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (lx *lexer) backup() {
	lx.pos -= lx.width
	// Correct newline count.
	if lx.width == 1 && lx.input[lx.pos] == '\n' {
		lx.line--
	}
}

// emit passes an item back to the client.
func (lx *lexer) emit(t tokenType) {
	value := lx.input[lx.start:lx.pos]
	lx.tokens <- token{t, lx.start, value, lx.startLine}
	lx.start = lx.pos
	lx.startLine = lx.line
}

// ignore skips over the pending input before this point.
func (lx *lexer) ignore() {
	currentVal := lx.input[lx.start:lx.pos]
	lx.line += strings.Count(currentVal, "\n") // todo: bug? all new lines should have been seen already
	lx.start = lx.pos
	lx.startLine = lx.line
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating lx.nextToken.
func (lx *lexer) errorf(format string, args ...interface{}) stateFn {
	lx.tokens <- token{tokenError, lx.start, fmt.Sprintf(format, args...), lx.startLine}
	return nil
}

// lex returns a new instance of lexer.
func lex(input string) *lexer {
	lx := &lexer{
		input:  input,
		tokens: make(chan token),
	}
	go lx.run()
	return lx
}

// run runs the state machine for the lexer.
func (lx *lexer) run() {
	for state := lexDocument; state != nil; {
		state = state(lx)
	}
	close(lx.tokens)

	/*
	lexInput ->

	 */
}

// lexDocument scans until tag is found (alphabetical character).
// can contain only whitespace and comments.
func lexDocument(lx *lexer) stateFn {
	switch rn := lx.next(); {
	case isWhitespace(rn):
		return lexWhitespace
	case rn == '-':
		return lexComment
	case unicode.IsLetter(rn):
		// todo: peek if doc tag
		return lexTagIdentifier
	}
}

// lexWhitespace scans a sequence of whitespace characters and ignores them.
// One whitespace has already been seen.
func lexWhitespace(lx *lexer) stateFn {
	for isWhitespace(lx.peek()) {
		lx.next()
	}
	lx.ignore()
	return lexDocument
}

// lexComment scans a single-line comment and ignores it.
// One dash symbol (part of comment marker) has already been seen.
func lexComment(lx *lexer) stateFn {
	if rn := lx.next(); rn != '-' {
		return lx.errorf("unexpected character %#U", rn)
	}
	// consume everything on that line
	for !isLineTerminator(lx.next()) {
	}
	lx.ignore()
	return lexDocument
}

func lexTagIdentifier(lx *lexer) stateFn {
	// todo: can contain letters, numbers and dashes
	// todo: can not end with dash

	var r rune
	for {
		r = lx.next()
		if !isAlphaNumeric(r) { // contains letter, numbers or dashes
			lx.backup()
			break
		}
	}

	if !isSpace(r) {
		return lx.errorf("invalid character %#U in tag name", r)
	}
}

// isWhitespace checks whether r is a whitespace (space/newline/tab...) character.
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

// isSpace checks whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isLineTerminator reports whether r is an end-of-line character.
func isLineTerminator(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or dash.
func isAlphaNumeric(r rune) bool {
	return r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
