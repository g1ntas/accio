package markup

import (
	"fmt"
	"testing"
)

var tokenName = map[tokenType]string {
	tokenError:       "error",
	tokenEOF:         "EOF",
	tokenSpace:       "space",
	tokenNewline:     "newline",
	tokenIdentifier:  "identifier",
	tokenLeftDelim:   "left delimiter",
	tokenRightDelim:  "right delimiter",
	tokenString:      "string",
	tokenBody:        "body",
	tokenAttrDeclare: "-",
	tokenAssign:      "=",
}

func (t tokenType) String() string {
	s := tokenName[t]
	if s == "" {
		return fmt.Sprintf("item%d", int(t))
	}
	return s
}

type lexTest struct {
	name string
	input string
	tokens []token
}

func mkToken(typ tokenType, text string) token {
	return token{
		typ: typ,
		val: text,
	}
}

var (
	tEOF = mkToken(tokenEOF, "")
	tAssign = mkToken(tokenAssign, "=")
	tAttrDeclare= mkToken(tokenAttrDeclare, "-")
	tSingleSpace = mkToken(tokenSpace, " ")
	tNewline = mkToken(tokenNewline, "\n")
	tBodyLeft = mkToken(tokenLeftDelim, "<<")
	tBodyRight = mkToken(tokenRightDelim, ">>")
)

var lexTests = []lexTest{
	{"empty", "", []token{tEOF}},
	{"whitespace", " \n\t", []token{tEOF}},
	{"comments", "-- this is a comment", []token{tEOF}},
	{"multiline comments", "-- line1\n--line2", []token{tEOF}},
	{"empty tag", "tag", []token{
		mkToken(tokenIdentifier, "tag"),
		tEOF,
	}},
	{"tag with single attribute", `tag -attr="value"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSingleSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenString, `"value"`),
		tEOF,
	}},
	{"tag with multiple attributes", `tag -attr1="1" -attr2="2"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSingleSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr1"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tSingleSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr2"),
		tAssign,
		mkToken(tokenString, `"2"`),
		tEOF,
	}},
	{"tag with inline body", `tag << test >>`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSingleSpace,
		tBodyLeft,
		mkToken(tokenBody, " test "),
		tBodyRight,
		tEOF,
	}},
	{"tag with multiline body", "tag << \t\n test \n>>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSingleSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenBody, " test "),
		tNewline,
		tBodyRight,
		tEOF,
	}},
	{"tag with attribute and body", `tag -a="1" << test >>`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSingleSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "a"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tSingleSpace,
		tBodyLeft,
		mkToken(tokenBody, " test "),
		tBodyRight,
		tEOF,
	}},
	// todo: multiple tags
	// todo: tag with empty attribute
	// todo: tag and attr separated by mutliple spaces
	// todo: attrs separated by mutliple spaces
	// todo: attr and body separated by mutliple spaces
	// todo: left and right delimiters within multi body (but not on newline)
	// todo: error: do not allow dash at the beginning of identifier (attr)
	// todo: error: do not allow dash at the end of identifier
	// todo: error: do not allow invalid characters in identifier (e.g. _)
	// todo: error: tag must start on newline
	// todo: error: do not allow invalid characters within input
	// todo: error: unclosed quotes
	// todo: error: attr without assignment
	// todo: error: attr without value
	// todo: error: attr unmatched delimiter
	// todo: error: whitespace before multiline right delimiter
	// todo: error: left delimiter on newline
	// todo: error: attribute on newline
}

// collect gathers the emitted tokens into a slice
func collect(t *lexTest, left, right string) (tokens []token) {
	lx := lex(t.name, t.input, left, right)
	for {
		token := lx.nextToken()
		tokens = append(tokens, token)
		if token.typ == tokenEOF || token.typ == tokenError {
			break
		}
	}
	return
}

func equal(t1, t2 []token, checkPos bool) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i := range t1 {
		if t1[i].typ != t2[i].typ {
			return false
		}
		if t1[i].val != t2[i].val {
			return false
		}
		if checkPos && t1[i].pos != t2[i].pos {
			return false
		}
		if checkPos && t1[i].line != t2[i].line {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		tokens := collect(&test, "", "")
		if !equal(tokens, test.tokens, false) {
			t.Errorf("%s:\ngot\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}