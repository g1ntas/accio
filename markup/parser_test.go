package markup

import (
	"fmt"
	"testing"
)

type parserTest struct {
	name string
	input string
	ast []*TagNode
}

func body(s string) *string{
	return &s
}

var astTests = []parserTest{
	{"empty", "", []*TagNode{}},
	{"whitespace", " \n\t", []*TagNode{}},
	{"comments", "# this is a comment", []*TagNode{}},
	{"multiline comments", "# line1\n#line2", []*TagNode{}},
	{"empty tag", "tag", []*TagNode{{Name: "tag"}}},
	{"reserved tag", "delimiters", []*TagNode{}},
	{"tag with single attribute", `tag -attr="value"`, []*TagNode{{
		Name: "tag",
		Attributes: map[string]*AttrNode{
			"attr": {Name: "attr", Value: "value"},
		},
	}}},
	{"tag with multiple attributes", `tag -attr1="1" -attr2="2"`, []*TagNode{{
		Name: "tag",
		Attributes: map[string]*AttrNode{
			"attr1": {Name: "attr1", Value: "1"},
			"attr2": {Name: "attr2", Value: "2"},
		},
	}}},
	{"tag with empty attribute", `tag -attr=""`, []*TagNode{{
		Name: "tag",
		Attributes: map[string]*AttrNode{
			"attr": {Name: "attr", Value: ""},
		},
	}}},
	{"tag with inline body", `tag << test >>`, []*TagNode{{
		Name: "tag",
		Body: body(" test "),
	}}},
	{"tag with multiline body", "tag <<\ntest\n>>", []*TagNode{{
		Name: "tag",
		Body: body("test"),
	}}},
	{"tag with attribute and body", `tag -a="1" <<test>>`, []*TagNode{{
		Name: "tag",
		Body: body("test"),
		Attributes: map[string]*AttrNode{
			"a": {Name: "a", Value: "1"},
		},
	}}},
	{"multiple empty tags", `tag1\ntag2`, []*TagNode{
		{Name: "tag1"},
		{Name: "tag2"},
	}},
	{"multiple tags with attributes and body",
		`tag -attr="1" <<this is first body>>`+"\n"+`tag -attr="2" <<this is second body>>`, []*TagNode{
		{
			Name: "tag",
			Body: body("this is first body"),
			Attributes: map[string]*AttrNode{
				"attr": {Name: "attr", Value: "1"},
			},
		},
		{
			Name: "tag",
			Body: body("this is second body"),
			Attributes: map[string]*AttrNode{
				"attr": {Name: "attr", Value: "2"},
			},
		},
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

var lexDelimTests = []lexTest{
	{"tag with inline body", `tag {{test}`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tCustomBodyLeft,
		mkToken(tokenBody, "test"),
		tCustomBodyRight,
		tEOF,
	}},
	{"tag with empty inline body", `tag {{}`, []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tCustomBodyLeft,
		mkToken(tokenBody, ""),
		tCustomBodyRight,
		tEOF,
	}},
	{"tag with multiline body", "tag {{ \t\ntest\n}", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tCustomBodyLeft,
		tNewline,
		mkToken(tokenBody, "test"),
		tNewline,
		tCustomBodyRight,
		tEOF,
	}},
	{"tag with empty multiline body", "tag {{ \t\n\n}", []token{
		mkToken(tokenIdentifier, "tag"),
		tSpace,
		tCustomBodyLeft,
		tNewline,
		mkToken(tokenBody, ""),
		tNewline,
		tCustomBodyRight,
		tEOF,
	}},
	// todo: test unicode delimiters
	// todo: test single char delimiters
	// todo: test single char left and multi char right delimiters
}

var (
	tCustomBodyLeft = mkToken(tokenLeftDelim, "{{")
	tCustomBodyRight = mkToken(tokenRightDelim, "}")
)


// Test bodies with different delimiters.
func TestDelims(t *testing.T) {
	for _, test := range lexDelimTests {
		tokens := collect(&test, "{{", "}")
		if !equal(tokens, test.tokens, false) {
			t.Errorf("%s:\ngot\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}

var lexPosTests = []lexTest{
	{"empty", "", []token{{tokenEOF, 0, "", 1}}},
	{"multiline tag", "tag -attr=\"1\" <<\n body\n>>", []token{
		{tokenIdentifier, 0, "tag", 1},
		{tokenSpace, 3, " ", 1},
		{tokenAttrDeclare, 4, "-", 1},
		{tokenIdentifier, 5, "attr", 1},
		{tokenAssign, 9, "=", 1},
		{tokenString, 10, `"1"`, 1},
		{tokenSpace, 13, " ", 1},
		{tokenLeftDelim, 14, "<<", 1},
		{tokenNewline, 16, "\n", 1},
		{tokenBody, 17, " body", 2},
		{tokenNewline, 22, "\n", 2},
		{tokenRightDelim, 23, ">>", 3},
		{tokenEOF, 25, "", 3},
	}},
	{"trimmed comments and whitespace", "# comment\n   \n\t\ntag1\n\n#comment2\ntag2", []token{
		{tokenIdentifier, 16, "tag1", 4},
		{tokenNewline, 20, "\n", 4},
		{tokenIdentifier, 32, "tag2", 7},
		{tokenEOF, 36, "", 7},
	}},
}

// Test token positions.
func TestPos(t *testing.T) {
	for _, test := range lexPosTests {
		tokens := collect(&test, "", "")
		if !equal(tokens, test.tokens, true) {
			t.Errorf("%s:\ngot\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
			if len(tokens) == len(test.tokens) {
				// Detailed print; avoid token.String() to expose the position value.
				for i := range tokens {
					if !equal(tokens[i:i+1], test.tokens[i:i+1], true) {
						tk1 := tokens[i]
						tk2 := test.tokens[i]
						t.Errorf("\n\t#%d: got {%v %d %q %d} expected {%v %d %q %d}",
							i, tk1.typ, tk1.pos, tk1.val, tk1.line, tk2.typ, tk2.pos, tk2.val, tk2.line)
					}
				}
			}
		}
	}
}

// todo: test that goroutine exits after error