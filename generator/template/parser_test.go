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
		"render predefined integer in template",
		`template <<{{var}}>>`,
		data{"var": 1},
		&model{Body: "1"},
	},
	{
		"render predefined string in template",
		`template <<{{var}}>>`,
		data{"var": "test"},
		&model{Body: "test"},
	},
	{
		"render predefined bool in template",
		`template <<{{var}}>>`,
		data{"var": true},
		&model{Body: "true"},
	},
	{
		"render predefined string list in template",
		`template <<{{#var}}{{.}}{{/var}}>>`,
		data{"var": []string{"a", "b"}},
		&model{Body: "ab"},
	},
	{
		"render integer var in template",
		`` +
			`variable -name="var" << 1 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "1"},
	},
	{
		"render string var in template",
		`` +
			`variable -name="var" << "test" >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "test"},
	},
	{
		"render boolean var in template",
		`` +
			`variable -name="var" << True >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "true"},
	},
	{
		"render float var in template",
		`` +
			`variable -name="var" << 1.5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "1.5"},
	},
	{
		"render null var in template",
		`` +
			`variable -name="var" << None >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: ""},
	},
	{
		"render list var in template",
		`` +
			`variable -name="var" << [1, "2", 3.1, True, None, ["a","b"]] >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "[1 2 3.1 true  [a b]]"},
	},
	{
		"render tuple var in template",
		`` +
			`variable -name="var" << (1, "2", 3.1, True, None, (1, 2)) >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "[1 2 3.1 true  [1 2]]"},
	},
	{
		"render dict var in template",
		`` +
			`variable -name="var" << {1: 1, 1.1: None, "a": "2", True: 3.1, None: True} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "map[1:1 1.1: None:true True:3.1 a:2]"},
	},
	{
		"render tuple indexed dict var in template",
		`` +
			`variable -name="var" << {("a",1,2.2,True,False,None,(1,2)): "test"} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "map[a 1 2.2 True False None 1 2:test]"},
	},
	{
		"pass int var to other var",
		`` +
			`variable -name="var1" << 1 >>					` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "int"},
	},
	{
		"pass float var to other var",
		`` +
			`variable -name="var1" << 1.1 >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "float"},
	},
	{
		"pass string var to other var",
		`` +
			`variable -name="var1" << "test" >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "string"},
	},
	{
		"pass null var to other var",
		`` +
			`variable -name="var1" << None >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "NoneType"},
	},
	{
		"pass bool var to other var",
		`` +
			`variable -name="var1" << True >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "bool"},
	},
	{
		"pass tuple var to other var",
		`` +
			`variable -name="var1" << (1,2) >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "tuple"},
	},
	{
		"pass list var to other var",
		`` +
			`variable -name="var1" << [1,2] >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "list"},
	},
	{
		"pass dict var to other var",
		`` +
			`variable -name="var1" << {1: "a"} >>			` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "dict"},
	},
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
		p, err := NewParser(test.data)
		if err != nil {
			t.Fatalf("%s:\nunexpected error: %v", test.name, err)
			return
		}
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
