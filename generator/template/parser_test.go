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

const (
	hasError = false
	noError  = true
	newline  = "\n"
)

func stringify(m *model) string {
	return fmt.Sprintf("{%q %q %v}", m.Filename, m.Body, m.Skip)
}

var parseTests = []struct {
	name  string
	input string
	data  data
	model *model
	ok    bool
}{
	{"empty", "", data{}, &model{}, noError},
	{"simple template", "template <<test>>", data{}, &model{Body: "test"}, noError},
	{"template with global variable", "template <<{{var}}>>", data{"var": "test"}, &model{Body: "test"}, noError},
	{
		"template with partial",
		`` +
			`partial -name="from" <<9>>` + newline +
			`partial -name="till" <<5>>` + newline +
			`template <<working {{> from}} to {{> till}}>>`,
		data{},
		&model{Body: "working 9 to 5"},
		noError,
	},
	{
		"partial with global variable",
		`` +
			`partial -name="partial" <<{{var}}>>` + newline +
			`template <<{{> partial}}>>`,
		data{"var": "test"},
		&model{Body: "test"},
		noError,
	},
	{
		"local variable",
		`` +
			`variable -name="var" <<` + newline +
			`	sum = 2 + 2			` + newline +
			`	return sum			` + newline +
			`>>						` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "4"},
		noError,
	},
	{
		"inline local variable",
		`` +
			`variable -name="var" << 5 + 5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "10"},
		noError,
	},
	{
		"overwrite global variable locally",
		`` +
			`variable -name="var" << 2 >>` + newline +
			`template <<{{var}}>>`,
		data{"var": 1},
		&model{Body: "2"},
		noError,
	},
	{
		"global variable in script",
		`` +
			`variable -name="local" << vars['global'] + 1 >>` + newline +
			`template <<{{local}}>>`,
		data{"global": 1},
		&model{Body: "2"},
		noError,
	},

	// verify data types are parsed correctly
	{
		"global integer var in template",
		`template <<{{var}}>>`,
		data{"var": 1},
		&model{Body: "1"},
		noError,
	},
	{
		"global string var in template",
		`template <<{{var}}>>`,
		data{"var": "test"},
		&model{Body: "test"},
		noError,
	},
	{
		"global bool var in template",
		`template <<{{var}}>>`,
		data{"var": true},
		&model{Body: "true"},
		noError,
	},
	{
		"global string list var in template",
		`template <<{{#var}}{{.}}{{/var}}>>`,
		data{"var": []string{"a", "b"}},
		&model{Body: "ab"},
		noError,
	},
	{
		"local integer var in template",
		`` +
			`variable -name="var" << 1 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "1"},
		noError,
	},
	{
		"local string var in template",
		`` +
			`variable -name="var" << "test" >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "test"},
		noError,
	},
	{
		"local boolean var in template",
		`` +
			`variable -name="var" << True >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "true"},
		noError,
	},
	{
		"local float var in template",
		`` +
			`variable -name="var" << 1.5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "1.5"},
		noError,
	},
	{
		"local null var in template",
		`` +
			`variable -name="var" << None >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: ""},
		noError,
	},
	{
		"local list var in template",
		`` +
			`variable -name="var" << [1, "2", 3.1, True, None, ["a","b"]] >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "[1 2 3.1 true  [a b]]"},
		noError,
	},
	{
		"local tuple var in template",
		`` +
			`variable -name="var" << (1, "2", 3.1, True, None, (1, 2)) >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "[1 2 3.1 true  [1 2]]"},
		noError,
	},
	{
		"local dict var in template",
		`` +
			`variable -name="var" << {1: 1, 1.1: None, "a": "2", True: 3.1, None: True} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "map[1:1 1.1: None:true True:3.1 a:2]"},
		noError,
	},
	{
		"local tuple indexed dict var in template",
		`` +
			`variable -name="var" << {("a",1,2.2,True,False,None,(1,2)): "test"} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "map[a 1 2.2 True False None 1 2:test]"},
		noError,
	},
	{
		"local int var in another local var",
		`` +
			`variable -name="var1" << 1 >>					` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "int"},
		noError,
	},
	{
		"local float var in another local var",
		`` +
			`variable -name="var1" << 1.1 >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "float"},
		noError,
	},
	{
		"local string var in another local var",
		`` +
			`variable -name="var1" << "test" >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "string"},
		noError,
	},
	{
		"local null var in another local var",
		`` +
			`variable -name="var1" << None >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "NoneType"},
		noError,
	},
	{
		"local bool var in another local var",
		`` +
			`variable -name="var1" << True >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "bool"},
		noError,
	},
	{
		"local tuple var in another local var",
		`` +
			`variable -name="var1" << (1,2) >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "tuple"},
		noError,
	},
	{
		"local list var in another local var",
		`` +
			`variable -name="var1" << [1,2] >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "list"},
		noError,
	},
	{
		"local dict var in another local var",
		`` +
			`variable -name="var1" << {1: "a"} >>			` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&model{Body: "dict"},
		noError,
	},
	{
		"overwrite previous local variable",
		`` +
			`variable -name="var" << "test" >>			` + newline +
			`variable -name="var" << vars['var'] >>		` + newline +
			`template <<{{var}}>>`,
		data{},
		&model{Body: "test"},
		noError,
	},

	// filename tag
	{
		"filename returns string",
		`filename << "test" >>`,
		data{},
		&model{Filename: "test"},
		noError,
	},
	{
		"filename returns null",
		`filename << None >>`,
		data{},
		&model{Filename: ""},
		noError,
	},
	{
		"filename with global var",
		`filename << vars['global'] >>`,
		data{"global": "test"},
		&model{Filename: "test"},
		noError,
	},
	{
		"filename with local var",
		`` +
			`variable -name="local" << "test" >>` + newline +
			`filename << vars['local'] >>`,
		data{},
		&model{Filename: "test"},
		noError,
	},
	{"filename returns boolean", `filename << True >>`, data{}, nil, hasError},
	{"filename returns int", `filename << 1 >>`, data{}, nil, hasError},
	{"filename returns float", `filename << 1.1 >>`, data{}, nil, hasError},
	{"filename returns dict", `filename << {} >>`, data{}, nil, hasError},
	{"filename returns list", `filename << [] >>`, data{}, nil, hasError},
	{"filename returns tuple", `filename << (1,) >>`, data{}, nil, hasError},

	// skipif tag
	{
		"skipif returns bool literal",
		`skipif << True >>`,
		data{},
		&model{Skip: true},
		noError,
	},
	{
		"skipif returns true value",
		`skipif << "true string" >>`,
		data{},
		&model{Skip: true},
		noError,
	},
	{
		"skipif with global var",
		`skipif << vars['global'] >>`,
		data{"global": 1},
		&model{Skip: true},
		noError,
	},
	{
		"skipif with local var",
		`` +
			`variable -name="local" << ["true"] >>` + newline +
			`skipif << vars['local'] >>`,
		data{},
		&model{Skip: true},
		noError,
	},
	/*{
		"skipif with local var",
		`template << {{{}} >>`,
		data{},
		&model{Skip: true},
		noError,
	},*/
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
		case err == nil && !test.ok:
			t.Errorf("%s\nexpected error; got none", test.name)
		case err != nil && test.ok:
			t.Errorf("%s:\nunexpected error: %v", test.name, err)
		case err != nil && !test.ok:
			continue // expected error, got one
		case stringify(model) != stringify(test.model):
			t.Errorf("%s:\ngot:\n\t%v\nexpected:\n\t%v", test.name, stringify(model), stringify(test.model))
		}
	}
}

// todo: write test to ensure thread safety, as same parser can be used to parse multiple files
