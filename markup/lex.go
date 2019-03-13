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
	tokenNewline                     // newline
	tokenIdentity                    // identity for tags and attributes

	tokenBool    // boolean literal
	tokenNumber  // number literal (including imaginary)
	tokenString  // string literal
	tokenBody    // raw tag body text between left and right delimiters

	tokenAttribute  // dash ('-') introducing an attribute declaration
	tokenAssign     // equals sign ('=') introducing an attribute assignment
)

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
	size      Pos        // size of last rune read from the input
	line      int        // 1+number of newlines seen
	startLine int        // start line of current token
	tokens    chan token // channel of scanned tokens
}

// value returns the string being scanned of current token
func (lx *lexer) value() string {
	return lx.input[lx.start:lx.pos]
}

// atEndOfFile checks whether there are any characters left to scan.
func (lx *lexer) atEndOfFile() bool {
	return int(lx.pos) >= len(lx.input)
}

// next returns and does consume the next rune in the input.
func (lx *lexer) next() rune {
	if lx.atEndOfFile() {
		lx.size = 0
		return eof
	}
	r, s := utf8.DecodeRuneInString(lx.input[lx.pos:])
	lx.size = Pos(s)
	lx.pos += lx.size
	if r == '\n' {
		lx.line++
	}
	return r
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
	lx.pos -= lx.size
	// Correct newline count.
	if lx.size == 1 && lx.input[lx.pos] == '\n' {
		lx.line--
	}
}

// emit passes an item back to the client.
func (lx *lexer) emit(t tokenType) {
	lx.tokens <- token{t, lx.start, lx.value(), lx.startLine}
	lx.start = lx.pos
	lx.startLine = lx.line
}

// ignore skips over the pending input before this point.
func (lx *lexer) ignore() {
	lx.line += strings.Count(lx.value(), "\n") // todo: bug? all new lines should have been seen already
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
}

// lexDocument scans until tag is found (alphabetical character).
// can contain only whitespace and comments.
func lexDocument(lx *lexer) stateFn {
	switch r := lx.next(); {
	case r == eof:
		lx.emit(tokenEOF)
		return nil
	case isWhitespace(r):
		return lexWhitespace
	case r == '-':
		return lexComment
	case isLetter(r):
		// todo: peek if doc tag
		return lexTagIdentifier
	}
}

// lexWhitespace scans a sequence of whitespace characters and ignores them.
// One whitespace has already been seen.
func lexWhitespace(lx *lexer) stateFn {
	for isWhitespace(lx.peek()) { // todo: optimize without peeking, and backup instead
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

// lexTagIdentifier scans a tag name.
// First letter has already been seen.
func lexTagIdentifier(lx *lexer) stateFn {
	var r rune
	for {
		r = lx.next()
		if !isLetter(r) && !isDigit(r) && r != '-' {
			lx.backup()
			break
		}
	}
	if !isWhitespace(r) || r != eof {
		return lx.errorf("invalid character %#U after the tag name, expected whitespace instead", r)
	}
	v := lx.value()
	if v[len(v)-1:] == "-" {
		return lx.errorf("invalid character %#U at the end of the tag name, expected letter or number instead", r)
	}
	lx.emit(tokenIdentity)
	return lexAfterTag
}

func lexAfterTag(lx *lexer) stateFn {
	switch r := lx.next(); {
	case isSpace(r):
		return lexSpace
	case r == '-':
		return lexAttribute
	case isLineTerminator(r):
		lx.emit(tokenNewline)
		return lexDocument
	case r == eof:
		return lexDocument
	case isPunctuation(r):
		return lexBodyLeftDelimiter
	default:
		return lx.errorf("unexpected character %#U", r)
	}
}

func lexSpace(lx *lexer) stateFn {
	for isSpace(lx.peek()) {
		lx.next()
	}
	lx.emit(tokenSpace)
	return lexAfterTag
}

func lexAttribute(lx *lexer) stateFn {
	return lexAfterTag
}

func lexBodyLeftDelimiter(lx *lexer) stateFn {
	return lexBody
}

func lexBody(lx *lexer) stateFn {
	return lexBodyRightDelimiter
}

func lexBodyRightDelimiter(lx *lexer) stateFn {
	return lexDocument
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

// isLetter checks whether r is an ASCII valid letter ([a-zA-Z]).
func isLetter(r rune) bool {
	return (r >= 65 && r <= 90) || (r >= 97 && r <= 122)
}

// isDigit checks whether r is an ASCII valid numeric digit ([0-9]).
func isDigit(r rune) bool {
	return r >= 48 && r <= 57
}

// isPunctuation checks whether r is an ASCII valid punctuation mark.
func isPunctuation(r rune) bool {
	return (r >= 33 && r <= 47) ||
		(r >= 58 && r <= 64) ||
		(r >= 91 && r <= 96) ||
		(r >= 123 && r <= 126)
}
