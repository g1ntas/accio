package markup

import (
	"flag"
	"fmt"
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
	{"multiple empty tags", `tag1\ntag2`, noError, []*TagNode{
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

	// errors fired in parser
}

func astEqual(ast1, ast2 []*TagNode) bool {
	if len(ast1) != len(ast2) {
		return false
	}
	for i, t1 := range ast1 {
		t2 := ast2[i]
		if t1.Name != t2.Name || t1.Body != t2.Body {
			return false
		}
		if len(t1.Attributes) != len(t2.Attributes) {
			return false
		}
		for name, attr1 := range t1.Attributes {
			attr2, ok := t2.Attributes[name]
			if !ok {
				return false
			}
			if attr1.Name != attr2.Name || attr1.Value != attr2.Value {
				return false
			}
		}
	}
	return true
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		ast, err := Parse("test.go", test.input, "", "")
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
		case !astEqual(test.ast, ast):
			// todo: represent expected and actual structures
			t.Errorf("%s=(%q): failed", test.name, test.input)
		}
	}
}