package manifest

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
)

type MockPrompter struct {
	input string
}

var _ Prompter = (*MockPrompter)(nil)

func (p *MockPrompter) Get(_, _ string, validator func(val string) error) (string, error) {
	err := validator(p.input)
	if err != nil {
		return "", err
	}
	return p.input, nil
}

func (p *MockPrompter) SelectOne(_, _ string, _ []string) (string, error) {
	return p.input, nil
}

func (p *MockPrompter) SelectMultiple(_, _ string, _ []string) ([]string, error) {
	return strings.Split(p.input, "\n"), nil
}

func (p *MockPrompter) Confirm(_, _ string) (bool, error) {
	return strconv.ParseBool(p.input)
}

var integerTests = []struct {
	name     string
	input    string
	expected interface{}
}{
	{"zero", "0", 0},
	{"short int", "1", 1},
	{"negative int", "-1", -1},
	{"positive signed int", "+1", 1},
	{"int64", "1234567890123456789", 1234567890123456789},
	{"too long int", "12345678901234567890", errIntOutOfRange},
	{"invalid int", "123-456", errNotInt},
	{"empty value", "", errNotInt},
}

func TestInteger(t *testing.T) {
	for _, test := range integerTests {
		t.Run(test.name, func(t *testing.T) {
			prompter := &MockPrompter{test.input}
			prompt := new(integer)
			val, err := prompt.Prompt(prompter)
			if _, isErr := test.expected.(error); isErr {
				require.Equal(t, test.expected, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, val)
			}
		})
	}
}
