// Inspired by Bob Pike's talk and text/template library.
// Parts of the source code were reused from golang/text/template/parse/lex.go.

package markup

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenType represents the type of the Token
type tokenType int

const (
	tokenError         tokenType = iota // error occurred; value is text of error
	tokenEOF                            // end of file
	tokenNewline                        // newline
	tokenIdentifier                     // identity for tags and attributes
	tokenLeftDelim                      // body opening delimiter
	tokenRightDelim                     // body closing delimiter
	tokenString                         // string literal
	tokenInlineBody                     // raw inline tag body text between left and right delimiters
	tokenMultilineBody                  // raw multiline tag body text between left and right delimiters
	tokenAttrDeclare                    // dash ('-') introducing an attribute declaration
	tokenAssign                         // equals sign ('=') introducing an attribute assignment
	tokenDelimiters                     // reserved tag keyword 'delimiters'
)

const eof = -1

// default body delimiters
const (
	leftDelimiter  = "<<"
	rightDelimiter = ">>"
)

// reserved tag names
var reservedTags = map[string]tokenType{
	"delimiters": tokenDelimiters,
}

// token represents a token returned from the scanner.
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
	case len(t.val) > 10:
		return fmt.Sprintf("%.10q...", t.val)
	}
	return fmt.Sprintf("%q", t.val)
}

// lexStateFn represents the state of the scanner as a function that returns the next state.
type lexStateFn func(*lexer) lexStateFn

// lexer holds the state of the scanner.
type lexer struct {
	input      string     // the string being scanned
	pos        Pos        // current position in the input
	start      Pos        // start position of current token
	size       Pos        // size of last rune read from the input
	line       int        // 1+number of newlines seen
	startLine  int        // start line of current token
	tokens     chan token // channel of scanned tokens
	leftDelim  string     // left delimiter for the body of the tag
	rightDelim string     // right delimiter for the body of the tag
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

// backup steps back one rune. Can only be called once per call of lx.next.
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
	lx.start = lx.pos
	lx.startLine = lx.line
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating lx.nextToken.
func (lx *lexer) errorf(format string, args ...interface{}) lexStateFn {
	lx.tokens <- token{tokenError, lx.start, fmt.Sprintf(format, args...), lx.startLine}
	return nil
}

// nextToken returns the next token from the input.
// Called by the Parser, not in the lexing goroutine.
func (lx *lexer) nextToken() token {
	return <-lx.tokens
}

// drain drains the output so the lexing goroutine will exit.
// Called by the Parser, not in the lexing goroutine.
func (lx *lexer) drain() {
	for range lx.tokens {
	}
}

// atString checks whether the next scanned run of characters is equal to provided string.
func (lx *lexer) atString(s string) bool {
	tail := lx.input[lx.pos-lx.size:]
	return strings.HasPrefix(tail, s)
}

// lex returns a new instance of lexer.
func lex(input, left, right string) *lexer {
	if left == "" {
		left = leftDelimiter
	}
	if right == "" {
		right = rightDelimiter
	}
	lx := &lexer{
		input:      input,
		leftDelim:  left,
		rightDelim: right,
		tokens:     make(chan token),
		line:       1,
		startLine:  1,
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
// Ignores whitespace and comments.
func lexDocument(lx *lexer) lexStateFn {
	switch r := lx.next(); {
	case r == eof:
		lx.emit(tokenEOF)
		return nil
	case isLetter(r):
		if lx.start != 0 {
			if prev := rune(lx.input[lx.start-1]); !isLineTerminator(prev) {
				return lx.errorf("misplaced character %#U, tag identifier must start on the newline", r)
			}
		}
		return lexTagIdentifier
	case isWhitespace(r):
		return lexWhitespace
	case r == '#':
		return lexComment
	default:
		return lx.errorf("invalid character %#U", r)
	}
}

// lexWhitespace scans a sequence of whitespace characters and ignores them.
// One whitespace has already been seen.
func lexWhitespace(lx *lexer) lexStateFn {
	for isWhitespace(lx.next()) {
	}
	lx.backup()
	lx.ignore()
	return lexDocument
}

// lexComment scans a single-line comment and ignores it.
// Comment identifier is already scanned so comment is already
// known to be present.
func lexComment(lx *lexer) lexStateFn {
	// consume everything on that line
	for {
		if r := lx.next(); isLineTerminator(r) || r == eof {
			break
		}
	}
	lx.ignore()
	return lexDocument
}

// lexTagIdentifier scans a tag full-name.
// First letter has already been seen.
func lexTagIdentifier(lx *lexer) lexStateFn {
	if !scanIdentifier(lx) {
		return nil
	}
	r := lx.peek()
	if !isSpace(r) && !isLineTerminator(r) && r != eof {
		return lx.errorf("invalid character %#U within tag identifier, space or newline expected", r)
	}
	if tokenType, exists := reservedTags[lx.value()]; exists {
		lx.emit(tokenType)
	} else {
		lx.emit(tokenIdentifier)
	}
	return lexAfterTag
}

// scanIdentifier scans identifier which can contain letters, numbers and dashes.
// First letter should already be scanned. Identifier can not end with dash.
func scanIdentifier(lx *lexer) bool {
	var r rune
	for {
		r = lx.next()
		if !isLetter(r) && !isDigit(r) && r != '-' {
			lx.backup()
			break
		}
	}
	v := lx.value()
	if len(v) > 0 && v[len(v)-1:] == "-" {
		lx.errorf("invalid character %#U at the end of the identifier", '-')
		return false
	}
	return true
}

// lexAfterTag scans inner tag (attributes and/or body).
func lexAfterTag(lx *lexer) lexStateFn {
	switch r := lx.next(); {
	case isSpace(r):
		return lexSpace
	case r == '-':
		lx.emit(tokenAttrDeclare)
		return lexAttributeName
	case isLineTerminator(r):
		lx.emit(tokenNewline)
		return lexDocument
	case r == eof:
		return lexDocument
	case lx.atString(lx.leftDelim):
		return lexBodyLeftDelimiter
	default:
		return lx.errorf("invalid character %#U", r)
	}
}

// lexSpace consumes a sequence of space characters.
// One space has already been seen.
func lexSpace(lx *lexer) lexStateFn {
	for isSpace(lx.next()) {
	}
	lx.backup()
	lx.ignore()
	return lexAfterTag
}

// lexAttribute scans an attribute full-name.
func lexAttributeName(lx *lexer) lexStateFn {
	if r := lx.next(); !isLetter(r) {
		return lx.errorf("invalid character %#U within attribute full-name, valid ascii letter expected", r)
	}
	if !scanIdentifier(lx) {
		return nil
	}
	lx.emit(tokenIdentifier)
	return lexAssignment
}

// lexAssignment scans an assignment character '='.
func lexAssignment(lx *lexer) lexStateFn {
	if r := lx.next(); r != '=' {
		return lx.errorf("invalid character %#U after the attribute, expected %#U instead", r, '=')
	}
	lx.emit(tokenAssign)
	return lexQuote
}

// lexQuote scans a quoted string (including quotes).
func lexQuote(lx *lexer) lexStateFn {
	if r := lx.next(); r != '"' {
		return lx.errorf("invalid character %#U after the attribute assignment, expected %#U instead", r, '"')
	}
Loop:
	for {
		switch lx.next() {
		case '"':
			break Loop
		case eof:
			return lx.errorf("unclosed attribute value, quote '\"' at the end expected")
		}
	}
	lx.emit(tokenString)
	return lexAfterTag
}

// lexBodyLeftDelimiter scans left (opening) delimiter which is known
// to be present. First char is already scanned. By default it's '<<',
// but can be changed by Parser.
func lexBodyLeftDelimiter(lx *lexer) lexStateFn {
	lx.pos += Pos(len(lx.leftDelim)) - lx.size
	lx.emit(tokenLeftDelim)
	return lexBody
}

// lexBody scans any text until a right (closing) delimiter is present.
// If newline is present after the left delimiter, scan multiline body.
func lexBody(lx *lexer) lexStateFn {
	for {
		switch r := lx.next(); {
		case isLineTerminator(r): // consume line terminator and all spaces
			lx.backup()
			lx.ignore()
			lx.next()
			lx.ignore()
			return lexMultilineBody
		case lx.atString(lx.rightDelim):
			lx.backup()
			lx.emit(tokenInlineBody)
			return lexBodyRightDelimiter
		case r == eof:
			return lx.errorf("unclosed tag body, ending delimiter \"%s\" expected at the end of body", lx.rightDelim)
		}
	}
}

// lexMultilineBody scans multiline text until a right delimiter is present on newline.
func lexMultilineBody(lx *lexer) lexStateFn {
	for {
		r := lx.next()
		if lx.atString("\n" + lx.rightDelim) {
			lx.backup()
			break
		}
		if r == eof {
			return lx.errorf("unclosed tag body, ending delimiter \"%s\" expected on newline at the end of the body", lx.rightDelim)
		}
	}
	lx.emit(tokenMultilineBody)
	lx.next()
	lx.ignore() // newline
	return lexBodyRightDelimiter
}

// lexBodyRightDelimiter scans right (closing) delimiter which is known
// to be present. By default it's '>>', but can be changed by Parser.
func lexBodyRightDelimiter(lx *lexer) lexStateFn {
	lx.pos += Pos(len(lx.rightDelim))
	lx.emit(tokenRightDelim)
	return lexNewlineAfterRightDelimiter
}

// lexNewlineAfterRightDelimiter scans newline token and ignores all
// space characters prior it. In case EOF is present, it successfully
// will finish the scanning.
func lexNewlineAfterRightDelimiter(lx *lexer) lexStateFn {
Loop:
	for {
		switch r := lx.next(); {
		case isSpace(r):
			continue Loop
		case isLineTerminator(r):
			lx.backup()
			lx.ignore() // ignore spaces
			break Loop
		case r == eof:
			lx.emit(tokenEOF)
			return nil
		default:
			return lx.errorf("invalid character %#U after right body delimiter", r)
		}
	}
	lx.next()
	lx.emit(tokenNewline)
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
	return r <= unicode.MaxASCII && unicode.IsLetter(r)
}

// isDigit checks whether r is an ASCII valid numeric digit ([0-9]).
func isDigit(r rune) bool {
	return r <= unicode.MaxASCII && unicode.IsDigit(r)
}
