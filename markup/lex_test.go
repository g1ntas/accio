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
	tEOF         = mkToken(tokenEOF, "")
	tAssign      = mkToken(tokenAssign, "=")
	tAttrDeclare = mkToken(tokenAttrDeclare, "-")
	tSpace       = mkToken(tokenSpace, " ")
	tNewline     = mkToken(tokenNewline, "\n")
	tBodyLeft    = mkToken(tokenLeftDelim, "<<")
	tBodyRight   = mkToken(tokenRightDelim, ">>")
)

var lexTests = []lexTest{
	{"empty", "", []token{tEOF}},
	{"whitespace", " \n\t", []token{tEOF}},
	{"comments", "# this is a comment", []token{tEOF}},
	{"multiline comments", "# line1\n#line2", []token{tEOF}},
	{"empty tag", "tag", []token{
		mkToken(tokenIdentifier, "tag"),
		tEOF,
	}},
	{"dash within tag", "tag-1", []token{
		mkToken(tokenIdentifier, "tag-1"),
		tEOF,
	}},
	{"tag with single attribute", `tag -attr="value"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenString, `"value"`),
		tEOF,
	}},
	{"tag with multiple attributes", `tag -attr1="1" -attr2="2"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr1"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr2"),
		tAssign,
		mkToken(tokenString, `"2"`),
		tEOF,
	}},
	{"tag with empty attribute value", `tag -attr=""`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenString, `""`),
		tEOF,
	}},
	{"tag and attribute separated by multiple spaces", "tag \t -attr=\"1\"", []token{
		mkToken(tokenIdentifier, "tag"),
		mkToken(tokenSpace, " \t "),
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tEOF,
	}},
	{"multiple attribute separated by multiple spaces", "tag -attr=\"1\" \t\t -attr=\"2\"", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenString, `"1"`),
		mkToken(tokenSpace, " \t\t "),
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenString, `"2"`),
		tEOF,
	}},
	{"spaces after empty tag", "tag \t\t", []token{
		mkToken(tokenIdentifier, "tag"),
		mkToken(tokenSpace, " \t\t"),
		tEOF,
	}},
	{"tag with inline body", `tag << test >>`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		mkToken(tokenBody, " test "),
		tBodyRight,
		tEOF,
	}},
	{"tag with multiline body", "tag << \t\n test \n>>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenBody, " test "),
		tNewline,
		tBodyRight,
		tEOF,
	}},
	{"tag with attribute and body", `tag -a="1" << test >>`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "a"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tSpace,
		tBodyLeft,
		mkToken(tokenBody, " test "),
		tBodyRight,
		tEOF,
	}},
	{"attribute and body separated by multiple spaces", "tag -a=\"1\" \t\t << test >>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "a"),
		tAssign,
		mkToken(tokenString, `"1"`),
		mkToken(tokenSpace, " \t\t "),
		tBodyLeft,
		mkToken(tokenBody, " test "),
		tBodyRight,
		tEOF,
	}},
	{"multiple empty tags", "tag1\ntag2", []token{
		mkToken(tokenIdentifier, "tag1"),
		tNewline,
		mkToken(tokenIdentifier, "tag2"),
		tEOF,
	}},
	{"multiple tags with attr", `tag1 -a="1"`+"\n"+`tag2`, []token{
		mkToken(tokenIdentifier, "tag1"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "a"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tNewline,
		mkToken(tokenIdentifier, "tag2"),
		tEOF,
	}},
	{"multiple tags with inline body", "tag1 <<body>>\ntag2", []token{
		mkToken(tokenIdentifier, "tag1"),
		tSpace,
		tBodyLeft,
		mkToken(tokenBody, "body"),
		tBodyRight,
		tNewline,
		mkToken(tokenIdentifier, "tag2"),
		tEOF,
	}},
	{"multiple tags with multiline body", "tag1 <<\nbody\n>>\ntag2", []token{
		mkToken(tokenIdentifier, "tag1"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenBody, "body"),
		tNewline,
		tBodyRight,
		tNewline,
		mkToken(tokenIdentifier, "tag2"),
		tEOF,
	}},
	{"multiple tags with attr and body", `tag1 -a="1" <<body>>`+"\n"+`tag2`, []token{
		mkToken(tokenIdentifier, "tag1"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "a"),
		tAssign,
		mkToken(tokenString, `"1"`),
		tSpace,
		tBodyLeft,
		mkToken(tokenBody, "body"),
		tBodyRight,
		tNewline,
		mkToken(tokenIdentifier, "tag2"),
		tEOF,
	}},
	{"spaces ignored after multiline body left delimiter", "tag << \t \ntest\n>>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenBody, "test"),
		tNewline,
		tBodyRight,
		tEOF,
	}},
	{"spaces ignored after inline body right delimiter", "tag <<test>> \t \n", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		mkToken(tokenBody, "test"),
		tBodyRight,
		tNewline,
		tEOF,
	}},
	{"spaces ignored after multiline body right delimiter", "tag <<\ntest\n>> \t \n", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenBody, "test"),
		tNewline,
		tBodyRight,
		tNewline,
		tEOF,
	}},
	{"delimiters within multiline body", "tag <<\n<<>>\n>>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenBody, "<<>>"),
		tNewline,
		tBodyRight,
		tEOF,
	}},

	// errors
	{"invalid character", "*", []token{
		mkToken(tokenError, "invalid character U+002A '*'"),
	}},
	{"invalid character within tag identifier", "t*g", []token{
		mkToken(tokenError, "invalid character U+002A '*' within tag identifier, space or newline expected"),
	}},
	{"dash at the start of the tag", "-tag", []token{
		mkToken(tokenError, "invalid character U+002D '-'"),
	}},
	{"dash at the end of the tag", "tag-", []token{
		mkToken(tokenError, "invalid character U+002D '-' at the end of the identifier"),
	}},
	{"tag must start on newline",  " tag", []token{
		mkToken(tokenError, "misplaced character U+0074 't', tag identifier must start on the newline"),
	}},
	{"invalid char at the start of attr name",  `tag -*="test"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenError, "invalid character U+002A '*' within attribute name, valid ascii letter expected"),
	}},
	{"dash at the start of attr name",  `tag --attr="test"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenError, "invalid character U+002D '-' within attribute name, valid ascii letter expected"),
	}},
	{"invalid char within attr name",  `tag -at*r="test"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "at"),
		mkToken(tokenError, "invalid character U+002A '*' after the attribute, expected U+003D '=' instead"),
	}},
	{"dash at the end of attr",  `tag -attr-="test"`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenError, "invalid character U+002D '-' at the end of the identifier"),
	}},
	{"attr without assignment",  `tag -attr`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		mkToken(tokenError, "invalid character U+FFFFFFFFFFFFFFFF after the attribute, expected U+003D '=' instead"),
	}},
	{"attr without value",  `tag -attr=`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenError, "invalid character U+FFFFFFFFFFFFFFFF after the attribute assignment, expected U+0022 '\"' instead"),
	}},
	{"unclosed attr quotes",  `tag -attr="`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tAttrDeclare,
		mkToken(tokenIdentifier, "attr"),
		tAssign,
		mkToken(tokenError, "unclosed attribute value, quote '\"' at the end expected"),
	}},
	{"delimiter after tag without space",  `tag<<>>`, []token{
		mkToken(tokenError, "invalid character U+003C '<' within tag identifier, space or newline expected"),
	}},
	{"unmatched inline body delimiter",  `tag << >`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		mkToken(tokenError, "unclosed tag body, ending delimiter \">>\" expected at the end of body"),
	}},
	{"unmatched multiline body delimiter",  "tag <<\n\n>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenError, "unclosed tag body, ending delimiter \">>\" expected on newline at the end of the body"),
	}},
	{"invalid char after body right delimiter",  "tag <<>>>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		mkToken(tokenBody, ""),
		tBodyRight,
		mkToken(tokenError, "invalid character U+003E '>' after right body delimiter"),
	}},
	{"whitespace before multiline right delimiter",  "tag <<\n\n\t>>", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tBodyLeft,
		tNewline,
		mkToken(tokenError, "unclosed tag body, ending delimiter \">>\" expected on newline at the end of the body"),
	}},
	{"left delimiter on newline",  "tag\n<<", []token{
		mkToken(tokenIdentifier, "tag"),
		tNewline,
		mkToken(tokenError, "invalid character U+003C '<'"),
	}},
	{"attr on newline",  "tag\n-attr", []token{
		mkToken(tokenIdentifier, "tag"),
		tNewline,
		mkToken(tokenError, "invalid character U+002D '-'"),
	}},
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