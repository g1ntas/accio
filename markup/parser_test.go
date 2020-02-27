package markup

import (
	"fmt"
	"testing"
)

func inlineBody(s string) *Body{
	return &Body{s, true}
}
func multilineBody(s string) *Body{
	return &Body{s, false}
}

const (
	noError = true
	hasError = false
)

var emptyAst []*TagNode
var parseTests = []struct {
	name string
	input string
	ok bool
	ast []*TagNode
} {
	{"empty", "", noError, emptyAst},
	{"whitespace", " \n\t", noError, emptyAst},
	{"comments", "# this is a comment", noError, emptyAst},
	{"multiline comments", "# line1\n#line2", noError, emptyAst},
	{"empty tag", "tag", noError, []*TagNode{{Name: "tag"}}},
	{"tag with single attribute", `tag -attr="value"`, noError, []*TagNode{{
		Name: "tag",
		Attributes: map[string]*AttrNode{
			"attr": {Name: "attr", Value: "value"},
		},
	}}},
	{"tag with multiple attributes", `tag -attr1="1" -attr2="2"`, noError, []*TagNode{{
		Name: "tag",
		Attributes: map[string]*AttrNode{
			"attr1": {Name: "attr1", Value: "1"},
			"attr2": {Name: "attr2", Value: "2"},
		},
	}}},
	{"tag with empty attribute", `tag -attr=""`, noError, []*TagNode{{
		Name: "tag",
		Attributes: map[string]*AttrNode{
			"attr": {Name: "attr", Value: ""},
		},
	}}},
	{"tag with inline body", `tag << test >>`, noError, []*TagNode{{
		Name: "tag",
		Body: inlineBody(" test "),
	}}},
	{"tag with multiline body", "tag <<\ntest\n>>", noError, []*TagNode{{
		Name: "tag",
		Body: multilineBody("test"),
	}}},
	{"tag with attribute and body", `tag -a="1" <<test>>`, noError, []*TagNode{{
		Name: "tag",
		Body: inlineBody("test"),
		Attributes: map[string]*AttrNode{
			"a": {Name: "a", Value: "1"},
		},
	}}},
	{"multiple empty tags", "tag1\ntag2", noError, []*TagNode{
		{Name: "tag1"},
		{Name: "tag2"},
	}},
	{"multiple tags with attributes and body",
		`tag -attr="1" <<this is first body>>` + "\n" + `tag -attr="2" <<this is second body>>`, noError, []*TagNode{
		{
			Name: "tag",
			Body: inlineBody("this is first body"),
			Attributes: map[string]*AttrNode{
				"attr": {Name: "attr", Value: "1"},
			},
		},
		{
			Name: "tag",
			Body: inlineBody("this is second body"),
			Attributes: map[string]*AttrNode{
				"attr": {Name: "attr", Value: "2"},
			},
		},
	}},

	// reserved delimiters tag
	{"delimiters tag | inline body with custom delimiters", `
delimiters -left="{" -right="}"
tag {body here}`, noError, []*TagNode{
		{
			Name: "tag",
			Body: inlineBody("body here"),
		},
	}},
	{"delimiters tag | multiline body with custom delimiters", `
delimiters -left="[[" -right="]]"
tag [[
  body here
]]`, noError, []*TagNode{
		{
			Name: "tag",
			Body: multilineBody("  body here"),
		},
	}},
	{"delimiters tag | no attrs", "delimiters", hasError, []*TagNode{}},
	{"delimiters tag | only left attr", `delimiters -left="{"`, hasError, []*TagNode{}},
	{"delimiters tag | only right attr", `delimiters -right="}"`, hasError, []*TagNode{}},
	{"delimiters tag | invalid attr", `delimiters -attr="test"`, hasError, []*TagNode{}},


	// lexer errors
	{"invalid character", "*", hasError, emptyAst},
	{"invalid character within tag identifier", "t*g", hasError, emptyAst},
	{"dash at the start of the tag", "-tag", hasError, emptyAst},
	{"dash at the end of the tag", "tag-", hasError, emptyAst},
	{"tag must start on newline",  " tag", hasError, emptyAst},
	{"invalid char at the start of attr full-name",  `tag -*="test"`, hasError, emptyAst},
	{"dash at the start of attr full-name",  `tag --attr="test"`, hasError, emptyAst},
	{"invalid char within attr full-name",  `tag -at*r="test"`, hasError, emptyAst},
	{"dash at the end of attr",  `tag -attr-="test"`, hasError, emptyAst},
	{"attr without assignment",  `tag -attr`, hasError, emptyAst},
	{"attr without value",  `tag -attr=`, hasError, emptyAst},
	{"unclosed attr quotes",  `tag -attr="`, hasError, emptyAst},
	{"delimiter after tag without space",  `tag<<>>`, hasError, emptyAst},
	{"unmatched inline body delimiter",  `tag << >`, hasError, emptyAst},
	{"unmatched multiline body delimiter",  "tag <<\n\n>", hasError, emptyAst},
	{"invalid char after body right delimiter",  "tag <<>>>", hasError, emptyAst},
	{"whitespace before multiline right delimiter",  "tag <<\n\n\t>>", hasError, emptyAst},
	{"left delimiter on newline",  "tag\n<<", hasError, emptyAst},
	{"attr on newline",  "tag\n-attr", hasError, emptyAst},
}

func astEqual(ast1, ast2 []*TagNode) bool {
	if len(ast1) != len(ast2) {
		return false
	}
	for i, t1 := range ast1 {
		t2 := ast2[i]
		str1 := t1.String()
		str2 := t2.String()
		if str1 != str2 {
			return false
		}
	}
	return true
}

func (n *AttrNode) String() string {
	return fmt.Sprintf("%s=%s", n.Name, n.Value)
}

func (n *TagNode) String() string {
	return fmt.Sprintf("{%s %v %v}", n.Name, n.Attributes, n.Body)
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		parser, err := Parse(test.input, "", "")
		switch {
		case err == nil && !test.ok:
			t.Errorf("%q: expected error; got none", test.name)
		case err != nil && test.ok:
			t.Errorf("%q: unexpected error: %v", test.name, err)
		case err != nil && !test.ok:
			// expected error, got one
		case !astEqual(test.ast, parser.Tags):
			t.Errorf("%s=(%q):\ngot\n\t%s\nexpected\n\t%s", test.name, test.input, parser.Tags, test.ast)
		}
	}
}