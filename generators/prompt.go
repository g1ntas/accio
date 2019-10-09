package generators

import (
	"errors"
	"strconv"
)

/*
Types:
`input` - normal input that takes string value
`integer`
`confirm` - confirm prompt, that returns `true` or `false`
`list` - list of values separated by `,`
`select` - prompt to select single value from given list
	* +options
`multi-select` - prompt to select multiple values from given list
	* +options
`file`-  valid file path, returns file stream
 */

const (
	PromptInput       = "input"
	PromptInteger     = "integer"
	PromptConfirm     = "confirm"
	PromptList        = "list"
	PromptSelect      = "select"
	PromptMultiSelect = "multi-select"
	PromptFile        = "file"
)

type InputValidatorFunc func(val string) error

type Prompter interface {
	Get(message, help string, validator InputValidatorFunc) (string, error)
	SelectOne(message, help string, options []string) (string, error)
	SelectMultiple(message, help string, options []string) (string, error)
	Confirm(message, help string) (bool, error)
}

type Prompt interface {
	Type() string
	Ask(prompter Prompter) (interface{}, error)
}

// Input
type Input struct {
	Msg  string `json:"message",yaml:"message"`
	Help string `json:"help",yaml:"help"`
}

func (p *Input) Type() string {
	return PromptInput
}

func (p *Input) Ask(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.Help, func(val string) error {
		return nil
	})
}


// Integer
type Integer struct {
	Msg  string `json:"message",yaml:"message"`
	Help string `json:"help",yaml:"help"`
}

func (p *Integer) Type() string {
	return PromptInteger
}

func (p *Integer) Ask(prompter Prompter) (interface{}, error) {
	val, err := prompter.Get(p.Msg, p.Help, func(val string) error {
		for i, r := range val {
			if r < '0' || r > '9' || (r == '-' && i != 0) {
				return errors.New("value is not an integer")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return strconv.Atoi(val)
}


// Confirm
type Confirm struct {
	Msg  string `json:"message",yaml:"message"`
	Help string `json:"help",yaml:"help"`
}

func (p *Confirm) Type() string {
	return PromptConfirm
}

func (p *Confirm) Ask(prompter Prompter) (interface{}, error) {
	return prompter.Confirm(p.Msg, p.Help)
}