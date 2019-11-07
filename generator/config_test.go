package generator

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"strings"
	"testing"
)

const (
	noError  = true
	hasError = false
)

var emptyGen = Generator{}

// conf is alias type for config to improve readability
type conf = map[string]interface{}

// strOfLen generates string of length n
func strOfLen(n int) string {
	return strings.Repeat("a", n)
}

// promptSignature returns Prompt data representation in string, used by generatorsEqual for comparison
func promptSignature(p Prompt) string {
	switch v := p.(type) {
	case *input:
		return fmt.Sprintf("%q %q", v.msg, v.help)
	case *integer:
		return fmt.Sprintf("%q %q", v.msg, v.help)
	case *confirm:
		return fmt.Sprintf("%q %q", v.msg, v.help)
	case *list:
		return fmt.Sprintf("%q %q", v.msg, v.help)
	case *choice:
		return fmt.Sprintf("%q %q %v", v.msg, v.help, v.options)
	case *multiChoice:
		return fmt.Sprintf("%q %q %v", v.msg, v.help, v.options)
	default:
		panic(fmt.Sprintf("Unknown Prompt: %s", p.kind()))
	}
}

// generatorsEqual compares two generators
func generatorsEqual(g1, g2 *Generator) bool {
	if g1.Name != g2.Name {
		return false
	}
	if g1.Description != g2.Description {
		return false
	}
	if g1.Help != g2.Help {
		return false
	}
	if len(g1.Prompts) != len(g2.Prompts) {
		return false
	}
	for k, prompt1 := range g1.Prompts {
		prompt2, ok := g2.Prompts[k]
		if !ok {
			return false
		}
		if promptSignature(prompt1) != promptSignature(prompt2) {
			return false
		}
	}
	return true
}

// string creates human-readable representation of Generator
func (g *Generator) string() string {
	// shorten description if too long
	var desc string
	if len(g.Description) > 10 {
		desc = fmt.Sprintf("%.10s...", g.Description)
	} else {
		desc = g.Description
	}
	// shorten help if too long
	var help string
	if len(g.Description) > 10 {
		help = fmt.Sprintf("%.10s...", g.Help)
	} else {
		help = g.Help
	}
	// stringify prompts in format [var]:[type]
	prompts := make([]string, len(g.Prompts))
	i := 0
	for k, p := range g.Prompts {
		prompts[i] = k+":"+p.kind()
		i++
	}
	return fmt.Sprintf("%q %q %q %v", g.Name, desc, help, prompts)
}

var configTests = []struct {
	name  string
	input map[string]interface{}
	gen   Generator
	ok    bool
}{
	// name
	{"valid name", conf{"name": "parse:test-command"}, Generator{Name: "parse:test-command"}, noError},
	{"name contains invalid characters", conf{"name": "."}, emptyGen, hasError},
	{"name starts with dash", conf{"name": "-a"}, emptyGen, hasError},
	{"name ends with dash", conf{"name": "a-"}, emptyGen, hasError},
	{"name starts with colon", conf{"name": ":a"}, emptyGen, hasError},
	{"name ends with colon", conf{"name": "a:"}, emptyGen, hasError},
	{"name longer than 64 characters", conf{"name": strOfLen(65)}, emptyGen, hasError},
	{"empty name", conf{"name": ""}, emptyGen, hasError},

	// description
	{"valid description", conf{"description": "abc"}, Generator{Description: "abc"}, noError},
	{"empty description", conf{"description": ""}, emptyGen, hasError},
	{"description longer than 128 characters", conf{"description": strOfLen(129)}, emptyGen, hasError},

	// help
	{"help", conf{"help": "abc"}, Generator{Help: "abc"}, noError},

	// prompts
	{
		"Prompt empty type",
		conf{"prompts": conf{"var": conf{"type": "", "message": "test"}}},
		Generator{},
		hasError,
	},
	{
		"Prompt invalid type",
		conf{"prompts": conf{"var": conf{"type": "invalid", "message": "test"}}},
		Generator{},
		hasError,
	},
	{
		"Prompt empty message",
		conf{"prompts": conf{"var": conf{"type": "input", "message": ""}}},
		Generator{},
		hasError,
	},
	{
		"Prompt message longer than 128 characters",
		conf{"prompts": conf{"var": conf{"type": "input", "message": strOfLen(125)}}},
		Generator{},
		hasError,
	},
	{
		"Prompt var name longer than 64 characters",
		conf{"prompts": conf{strOfLen(65): conf{"type": "input", "message": "test"}}},
		Generator{},
		hasError,
	},
	{
		"Prompt with valid var name",
		conf{"prompts": conf{"_Var_1": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"_Var_1": &input{base{msg: "test"}}}},
		noError,
	},
	{
		"Prompt with var name containing hyphen",
		conf{"prompts": conf{"test-var": conf{"type": "input", "message": "test"}}},
		Generator{},
		hasError,
	},
	{
		"Prompt with var name starting with digit",
		conf{"prompts": conf{"0var": conf{"type": "input", "message": "test"}}},
		Generator{},
		hasError,
	},
	{
		"Prompt with var name containing non-ascii characters",
		conf{"prompts": conf{"vaƒr": conf{"type": "input", "message": "test"}}},
		Generator{},
		hasError,
	},
	{
		"Prompt type input",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &input{base{msg: "test"}}}},
		noError,
	},
	{
		"Prompt help",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test", "help": "abc"}}},
		Generator{Prompts: PromptMap{"var": &input{base{msg: "test", help: "abc"}}}},
		noError,
	},
	{
		"Prompt help longer than 512 characters",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test", "help": strOfLen(513)}}},
		Generator{},
		hasError,
	},
	{
		"Prompt type integer",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &integer{base{msg: "test"}}}},
		noError,
	},
	{
		"Prompt type confirm",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &confirm{base{msg: "test"}}}},
		noError,
	},
	{
		"Prompt type list",
		conf{"prompts": conf{"var": conf{"type": "list", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &confirm{base{msg: "test"}}}},
		noError,
	},
	{
		"Prompt type choice",
		conf{"prompts": conf{"var": conf{
			"type": "choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Prompts: PromptMap{"var":
		&choice{
			base{msg: "test"},
			[]string{"a", "b"},
		},
		}},
		noError,
	},
	{
		"Prompt type multi choice",
		conf{"prompts": conf{"var": conf{
			"type": "multi-choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Prompts: PromptMap{"var":
		&multiChoice{
			base{msg: "test"},
			[]string{"a", "b"},
		},
		}},
		noError,
	},
}

func TestConfigReading(t *testing.T) {
	for _, test := range configTests {
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(test.input); err != nil {
			t.Fatalf("%s:\nfailed to encode config data to toml format\nreason: %v", test.name, err)
			return
		}
		gen := &Generator{Prompts: make(PromptMap)}
		err := gen.readConfig([]byte(buf.String()))
		switch {
		case err == nil && !test.ok:
			t.Errorf("%s: expected error; got none", test.name)
		case err != nil && test.ok:
			t.Errorf("%s: unexpected error: %v", test.name, err)
		case err != nil && !test.ok:
			continue // expected error, got one
		case !generatorsEqual(gen, &test.gen):
			t.Errorf("%s:\ngot:\n\t%v\nexpected:\n\t%v", test.name, gen.string(), test.gen.string())
		}
	}
}