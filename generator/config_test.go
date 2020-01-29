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
		return fmt.Sprintf("%q %q", v.Msg, v.HelpText)
	case *integer:
		return fmt.Sprintf("%q %q", v.Msg, v.HelpText)
	case *confirm:
		return fmt.Sprintf("%q %q", v.Msg, v.HelpText)
	case *list:
		return fmt.Sprintf("%q %q", v.Msg, v.HelpText)
	case *choice:
		return fmt.Sprintf("%q %q %v", v.Msg, v.HelpText, v.options)
	case *multiChoice:
		return fmt.Sprintf("%q %q %v", v.Msg, v.HelpText, v.options)
	default:
		panic(fmt.Sprintf("Unknown Prompt: %s", p.kind()))
	}
}

// generatorsEqual compares two generators
func generatorsEqual(g1, g2 *Generator) bool {
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
	// shorten help if too long
	var help string
	if len(g.Help) > 10 {
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
	return fmt.Sprintf("%q %v", help, prompts)
}

var configTests = []struct {
	name  string
	input map[string]interface{}
	gen   Generator
	ok    bool
}{
	// help
	{"help", conf{"name": "a", "help": "abc"}, Generator{Help: "abc"}, noError},

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
		Generator{Prompts: PromptMap{"_Var_1": &input{Base{Msg: "test"}}}},
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
		Generator{Prompts: PromptMap{"var": &input{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt help",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test", "help": "abc"}}},
		Generator{Prompts: PromptMap{"var": &input{Base{Msg: "test", HelpText: "abc"}}}},
		noError,
	},
	{
		"Prompt type integer",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &integer{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type confirm",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &confirm{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type list",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "list", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &confirm{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type choice",
		conf{"name": "a", "prompts": conf{"var": conf{
			"type":    "choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Prompts: PromptMap{"var":
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
		Generator{Prompts: PromptMap{"var":
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
