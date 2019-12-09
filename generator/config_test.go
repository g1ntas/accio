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
		return fmt.Sprintf("%q %q", v.Msg, v.Help)
	case *integer:
		return fmt.Sprintf("%q %q", v.Msg, v.Help)
	case *confirm:
		return fmt.Sprintf("%q %q", v.Msg, v.Help)
	case *list:
		return fmt.Sprintf("%q %q", v.Msg, v.Help)
	case *choice:
		return fmt.Sprintf("%q %q %v", v.Msg, v.Help, v.options)
	case *multiChoice:
		return fmt.Sprintf("%q %q %v", v.Msg, v.Help, v.options)
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
		prompts[i] = k + ":" + p.kind()
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
	{"valid description", conf{"name": "a", "description": "abc"}, Generator{Name: "a", Description: "abc"}, noError},
	{"description longer than 128 characters", conf{"name": "a", "description": strOfLen(129)}, emptyGen, hasError},

	// help
	{"help", conf{"name": "a", "help": "abc"}, Generator{Name: "a", Help: "abc"}, noError},

	// prompts
	{
		"Prompt empty type",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt invalid type",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "invalid", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt empty message",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": ""}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt message longer than 128 characters",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": strOfLen(129)}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt var name longer than 64 characters",
		conf{"name": "a", "prompts": conf{strOfLen(65): conf{"type": "input", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt with valid var name",
		conf{"name": "a", "prompts": conf{"_Var_1": conf{"type": "input", "message": "test"}}},
		Generator{Name: "a", Prompts: PromptMap{"_Var_1": &input{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt with var name starting with digit",
		conf{"name": "a", "prompts": conf{"0var": conf{"type": "input", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt with var name containing invalid characters",
		conf{"name": "a", "prompts": conf{"test-var": conf{"type": "input", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt type input",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Name: "a", Prompts: PromptMap{"var": &input{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt help",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test", "help": "abc"}}},
		Generator{Name: "a", Prompts: PromptMap{"var": &input{Base{Msg: "test", Help: "abc"}}}},
		noError,
	},
	{
		"Prompt type integer",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Name: "a", Prompts: PromptMap{"var": &integer{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type confirm",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Name: "a", Prompts: PromptMap{"var": &confirm{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type list",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "list", "message": "test"}}},
		Generator{Name: "a", Prompts: PromptMap{"var": &confirm{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type choice",
		conf{"name": "a", "prompts": conf{"var": conf{
			"type":    "choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Name: "a", Prompts: PromptMap{"var":
		&choice{
			Base{Msg: "test"},
			[]string{"a", "b"},
		},
		}},
		noError,
	},
	{
		"Prompt 'choice' without options",
		conf{"name": "a", "prompts": conf{"var": conf{
			"type":    "choice",
			"message": "test",
		}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt type multi choice",
		conf{"name": "a", "prompts": conf{"var": conf{
			"type":    "multi-choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Name: "a", Prompts: PromptMap{"var":
		&multiChoice{
			Base{Msg: "test"},
			[]string{"a", "b"},
		},
		}},
		noError,
	},
	{
		"Prompt 'multi-choice' without options",
		conf{"name": "a", "prompts": conf{"var": conf{
			"type":    "multi-choice",
			"message": "test",
		}}},
		emptyGen,
		hasError,
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
