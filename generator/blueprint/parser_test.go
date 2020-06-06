package blueprint

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type data = map[string]interface{}

const (
	hasError = false
	noError  = true
	newline  = "\n"
)

var parseTests = []struct {
	name      string
	input     string
	data      data
	blueprint *blueprint
	ok        bool
}{
	{"empty", "", data{}, &blueprint{}, noError},
	{"simple template", "template <<test>>", data{}, &blueprint{Body: "test"}, noError},
	{"template with global variable", "template <<{{var}}>>", data{"var": "test"}, &blueprint{Body: "test"}, noError},
	{
		"template with partial",
		`` +
			`partial -name="from" <<9>>` + newline +
			`partial -name="till" <<5>>` + newline +
			`template <<working {{> from}} to {{> till}}>>`,
		data{},
		&blueprint{Body: "working 9 to 5"},
		noError,
	},
	{
		"template with nested partial",
		`` +
			`partial -name="workTime" <<9 to 5>>` + newline +
			`partial -name="sentence" <<working {{> workTime}}>>` + newline +
			`template <<{{sentence}}>>`,
		data{},
		&blueprint{Body: "working 9 to 5"},
		noError,
	},
	{
		"partial with global variable",
		`` +
			`partial -name="partial" <<{{var}}>>` + newline +
			`template <<{{> partial}}>>`,
		data{"var": "test"},
		&blueprint{Body: "test"},
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
		&blueprint{Body: "4"},
		noError,
	},
	{
		"inline local variable",
		`` +
			`variable -name="var" << 5 + 5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "10"},
		noError,
	},
	{
		"overwrite global variable locally",
		`` +
			`variable -name="var" << 2 >>` + newline +
			`template <<{{var}}>>`,
		data{"var": 1},
		&blueprint{Body: "2"},
		noError,
	},
	{
		"global variable in script",
		`` +
			`variable -name="local" << vars['global'] + 1 >>` + newline +
			`template <<{{local}}>>`,
		data{"global": 1},
		&blueprint{Body: "2"},
		noError,
	},

	// verify data types are parsed correctly
	{
		"global integer var in template",
		`template <<{{var}}>>`,
		data{"var": 1},
		&blueprint{Body: "1"},
		noError,
	},
	{
		"global string var in template",
		`template <<{{var}}>>`,
		data{"var": "test"},
		&blueprint{Body: "test"},
		noError,
	},
	{
		"global bool var in template",
		`template <<{{var}}>>`,
		data{"var": true},
		&blueprint{Body: "true"},
		noError,
	},
	{
		"global string list var in template",
		`template <<{{#var}}{{.}}{{/var}}>>`,
		data{"var": []string{"a", "b"}},
		&blueprint{Body: "ab"},
		noError,
	},
	{
		"local integer var in template",
		`` +
			`variable -name="var" << 1 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "1"},
		noError,
	},
	{
		"local string var in template",
		`` +
			`variable -name="var" << "test" >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "test"},
		noError,
	},
	{
		"local boolean var in template",
		`` +
			`variable -name="var" << True >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "true"},
		noError,
	},
	{
		"local float var in template",
		`` +
			`variable -name="var" << 1.5 >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "1.5"},
		noError,
	},
	{
		"local null var in template",
		`` +
			`variable -name="var" << None >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: ""},
		noError,
	},
	{
		"local list var in template",
		`` +
			`variable -name="var" << [1, "2", 3.1, True, None, ["a","b"]] >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "[1 2 3.1 true  [a b]]"},
		noError,
	},
	{
		"local tuple var in template",
		`` +
			`variable -name="var" << (1, "2", 3.1, True, None, (1, 2)) >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "[1 2 3.1 true  [1 2]]"},
		noError,
	},
	{
		"local dict var in template",
		`` +
			`variable -name="var" << {1: 1, 1.1: None, "a": "2", True: 3.1, None: True} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "map[1:1 1.1: None:true True:3.1 a:2]"},
		noError,
	},
	{
		"local tuple indexed dict var in template",
		`` +
			`variable -name="var" << {("a",1,2.2,True,False,None,(1,2)): "test"} >>` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "map[a 1 2.2 True False None 1 2:test]"},
		noError,
	},
	{
		"local int var in another local var",
		`` +
			`variable -name="var1" << 1 >>					` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "int"},
		noError,
	},
	{
		"local float var in another local var",
		`` +
			`variable -name="var1" << 1.1 >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "float"},
		noError,
	},
	{
		"local string var in another local var",
		`` +
			`variable -name="var1" << "test" >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "string"},
		noError,
	},
	{
		"local null var in another local var",
		`` +
			`variable -name="var1" << None >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "NoneType"},
		noError,
	},
	{
		"local bool var in another local var",
		`` +
			`variable -name="var1" << True >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "bool"},
		noError,
	},
	{
		"local tuple var in another local var",
		`` +
			`variable -name="var1" << (1,2) >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "tuple"},
		noError,
	},
	{
		"local list var in another local var",
		`` +
			`variable -name="var1" << [1,2] >>				` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "list"},
		noError,
	},
	{
		"local dict var in another local var",
		`` +
			`variable -name="var1" << {1: "a"} >>			` + newline +
			`variable -name="var2" << type(vars['var1']) >>	` + newline +
			`template <<{{var2}}>>`,
		data{},
		&blueprint{Body: "dict"},
		noError,
	},
	{
		"overwrite previous local variable",
		`` +
			`variable -name="var" << "test" >>			` + newline +
			`variable -name="var" << vars['var'] >>		` + newline +
			`template <<{{var}}>>`,
		data{},
		&blueprint{Body: "test"},
		noError,
	},

	// filename tag
	{
		"filename returns string",
		`filename << "test" >>`,
		data{},
		&blueprint{Filename: "test"},
		noError,
	},
	{
		"filename returns null",
		`filename << None >>`,
		data{},
		&blueprint{Filename: ""},
		noError,
	},
	{
		"filename with global var",
		`filename << vars['global'] >>`,
		data{"global": "test"},
		&blueprint{Filename: "test"},
		noError,
	},
	{
		"filename with local var",
		`` +
			`variable -name="local" << "test" >>` + newline +
			`filename << vars['local'] >>`,
		data{},
		&blueprint{Filename: "test"},
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
		&blueprint{Skip: true},
		noError,
	},
	{
		"skipif returns true value",
		`skipif << "true string" >>`,
		data{},
		&blueprint{Skip: true},
		noError,
	},
	{
		"skipif with global var",
		`skipif << vars['global'] >>`,
		data{"global": 1},
		&blueprint{Skip: true},
		noError,
	},
	{
		"skipif with local var",
		`` +
			`variable -name="local" << ["true"] >>` + newline +
			`skipif << vars['local'] >>`,
		data{},
		&blueprint{Skip: true},
		noError,
	},
}

func TestParsing(t *testing.T) {
	for _, test := range parseTests {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewParser(test.data)
			require.NoError(t, err)

			bp, err := p.Parse([]byte(test.input))
			switch {
			case !test.ok:
				assert.Error(t, err)
			case test.ok:
				assert.NoError(t, err)
			default:
				assert.Equal(t, test.blueprint, bp)
			}
		})
	}
}

var errorTests = []struct {
	name  string
	input string
	tag   string
	line  int
}{
	{"mustache error", "template <<\n{{}\n>>", "template", 2},
	{"inline mustache error", "template << {{} >>", "template", 1},
	{"inline starlark error with line", `filename << © >>`, "filename", 1},
	{"starlark error with line", "filename <<\n\treturn ©\n>>", "filename", 2},
	{"starlark error without line", "\n\n\nfilename <<\n\treturn 1/0\n>>", "filename", 4},
	{"starlark error list", "filename <<\n\treturn undefined\n>>", "filename", 2},
}

func TestErrors(t *testing.T) {
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewParser(data{})
			require.NoError(t, err)

			_, err = p.Parse([]byte(test.input))
			require.Error(t, err)
			require.IsType(t, &ParseError{}, err)

			e := err.(*ParseError)
			assert.Equal(t, test.tag, e.Tag)
			assert.Equal(t, test.line, e.Line)
		})
	}
}
