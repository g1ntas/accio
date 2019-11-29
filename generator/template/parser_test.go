package template

import (
	"fmt"
	"github.com/g1ntas/accio/generator"
	"testing"
)

// render simple template
// render template with predefined variable

type model = generator.Template
type data = map[string]interface{}

const newline = "\n"

func stringify(m *model) string {
	return fmt.Sprintf("{%q %q %v}", m.Filename, m.Body, m.Skip)
}

var parseTests = []struct {
	name  string
	input string
	data  data
	model *model
}{
	{"empty", "", data{}, &model{}},
	{"simple template", "template <<test>>", data{}, &model{Body: "test"}},
	{"template with predefined variable", "template <<{{var}}>>", data{"var": "test"}, &model{Body: "test"}},
	{
		"template with partial",
		`` +
			`partial -name="from" <<9>>` + newline +
			`partial -name="till" <<5>>` + newline +
			`template <<working {{> from}} to {{> till}}>>`,
		data{},
		&model{Body: "working 9 to 5"},
	},
	{
		"partial with predefined variable",
		`` +
			`partial -name="partial" <<{{var}}>>` + newline +
			`template <<{{> partial}}>>`,
		data{"var": "test"},
		&model{Body: "test"},
	},
	{
		"simple variable",
		`` +
			`variable -name="var" <<` + newline +
			`	sum = 2 + 2			` + newline +
			`	return sum			` + newline +
			`>>						` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "4"},
	},
	{
		"inline variable",
		`` +
			`variable -name="var" << 5 + 5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "10"},
	},
	{
		"overwrite predefined variable",
		`` +
			`variable -name="var" << 2 >>` + newline +
			`template <<{{var}}>>`,
		data{"var": 1},
		&model{Body: "2"},
	},
	{
		"use predefined variable inside script",
		`` +
			`variable -name="var" << vars['var'] + 1 >>` + newline +
			`template <<{{var}}>>`,
		data{"var": 1},
		&model{Body: "2"},
	},

	// verify data types are parsed correctly
	{
		"interpret predefined integer in template",
		`template <<{{var}}>>`,
		data{"var": 1},
		&model{Body: "1"},
	},
	{
		"interpret predefined string in template",
		`template <<{{var}}>>`,
		data{"var": "test"},
		&model{Body: "test"},
	},
	{
		"interpret predefined bool in template",
		`template <<{{var}}>>`,
		data{"var": true},
		&model{Body: "true"},
	},
	{
		"interpret predefined string list in template",
		`template <<{{#var}}{{.}}{{/var}}>>`,
		data{"var": []string{"a", "b"}},
		&model{Body: "ab"},
	},
	{
		"interpret integer from script in template",
		`` +
			`variable -name="var" << 1 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "1"},
	},
	{
		"interpret string from script in template",
		`` +
			`variable -name="var" << "test" >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "test"},
	},
	{
		"interpret boolean from script in template",
		`` +
			`variable -name="var" << True >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "true"},
	},
	{
		"interpret float from script in template",
		`` +
			`variable -name="var" << 1.5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "1.5"},
	},
	{
		"interpret null from script in template",
		`` +
			`variable -name="var" << None >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: ""},
	},
	{
		"interpret list from script in template",
		`` +
			`variable -name="var" << [1, "2", 3.1, True, None, ["a","b"]] >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "[1 2 3.1 true  [a b]]"},
	},
	{
		"interpret tuple from script in template",
		`` +
			`variable -name="var" << (1, "2", 3.1, True, None, (1, 2)) >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "[1 2 3.1 true  [1 2]]"},
	},
	{
		"interpret dict from script in template",
		`` +
			`variable -name="var" << {1: 1, 1.1: None, "a": "2", True: 3.1, None: True} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "map[1:1 1.1: None:true True:3.1 a:2]"},
	},
	{
		"interpret tuple indexed dict from script in template",
		`` +
			`variable -name="var" << {("a",1,2.2,True,False,None,(1,2)): "test"} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "map[a 1 2.2 True False None 1 2:test]"},
	},

	// todo: test local int variable is correctly passed to other var
	// todo: test local float variable is correctly passed to other var
	// todo: test local bool variable is correctly passed to other var
	// todo: test local null variable is correctly passed to other var
	// todo: test local tuple variable is correctly passed to other var
	// todo: test local list variable is correctly passed to other var
	// todo: test local dict variable is correctly passed to other var

	// todo: test filename tag works
	// todo: test that filename script is executed
	// todo: test filename can not return any other type than string
	// todo: test that global variables can be used in filename tag
	// todo: test that local variables can be used in filename tag

	// todo: test skipif tag works
	// todo: test that skipif script is executed
	// todo: test skipif can not return any other type than boolean
	// todo: test that global variables can be used in skipif tag
	// todo: test that local variables can be used in skipif tag

}

func TestParsing(t *testing.T) {
	for _, test := range parseTests {
		p := NewParser(test.data)
		model, err := p.Parse([]byte(test.input))
		switch {
		case err != nil:
			t.Errorf("%s:\nunexpected error: %v", test.name, err)
		case stringify(model) != stringify(test.model):
			t.Errorf("%s:\ngot:\n\t%v\nexpected:\n\t%v", test.name, stringify(model), stringify(test.model))
		}
	}
}

// todo: write test to ensure thread safety, as same parser can be used to parse multiple files
