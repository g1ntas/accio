package markup

import (
	"flag"
	"fmt"
	"strings"
	"testing"
)

var debug = flag.Bool("debug", false, "show the errors produced by the main tests")

func body(s string) *string{
	return &s
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
	{"reserved tag", "delimiters", noError, []*TagNode{}},
	// todo: reserved tag with attrs
	// todo: reserved tag and normal tag
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
		Body: body(" test "),
	}}},
	{"tag with multiline body", "tag <<\ntest\n>>", noError, []*TagNode{{
		Name: "tag",
		Body: body("test"),
	}}},
	{"tag with attribute and body", `tag -a="1" <<test>>`, noError, []*TagNode{{
		Name: "tag",
		Body: body("test"),
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

	// errors fired in lexer
	{"invalid character", "*", hasError, emptyAst},
	{"invalid character within tag identifier", "t*g", hasError, emptyAst},
	{"dash at the start of the tag", "-tag", hasError, emptyAst},
	{"dash at the end of the tag", "tag-", hasError, emptyAst},
	{"tag must start on newline",  " tag", hasError, emptyAst},
	{"invalid char at the start of attr name",  `tag -*="test"`, hasError, emptyAst},
	{"dash at the start of attr name",  `tag --attr="test"`, hasError, emptyAst},
	{"invalid char within attr name",  `tag -at*r="test"`, hasError, emptyAst},
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

	// errors fired in Parser
}

func astEqual(ast1, ast2 []*TagNode) bool {
	if len(ast1) != len(ast2) {
		return false
	}
	for i, t1 := range ast1 {
		t2 := ast2[i]
		if t1.String() != t2.String() {
			return false
		}
	}
	return true
}

func (n *AttrNode) String() string {
	return fmt.Sprintf("%s=%s", n.Name, n.Value)
}

func (n *TagNode) String() string {
	var attrs []string
	for _, attr := range n.Attributes {
		attrs = append(attrs, attr.String())
	}
	var body string
	if n.Body != nil {
		// (%.10q...)
		body = fmt.Sprintf(" (%s) ", *n.Body)
	}
	if len(attrs) > 0 {
 		return fmt.Sprintf("%s%s (%s)", n.Name, body, strings.Join(attrs, ","))
	}
	return fmt.Sprintf("%s%s ", n.Name, body)
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		parser, err := Parse("test.go", test.input, "", "")
		switch {
		case err == nil && !test.ok:
			t.Errorf("%q: expected error; got none", test.name)
		case err != nil && test.ok:
			t.Errorf("%q: unexpected error: %v", test.name, err)
		case err != nil && !test.ok:
			// expected error, got one
			if *debug {
				fmt.Printf("%s: %s\n\t%s\n", test.name, test.input, err)
			}
		case !astEqual(test.ast, parser.Tags):
			// todo: represent expected and actual structures
			t.Errorf("%s=(%q): %v", test.name, test.input, parser.Tags)
		}
	}
}