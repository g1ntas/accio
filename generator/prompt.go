package generator

import (
	"errors"
	"strconv"
)

type Prompter interface {
	Get(message, help string, validator func(val string) error) (string, error)
	SelectOne(message, help string, options []string) (string, error)
	SelectMultiple(message, help string, options []string) ([]string, error)
	Confirm(message, help string) (bool, error)
}

const (
	promptInput       = "input"
	promptInteger     = "integer"
	promptConfirm     = "confirm"
	promptChoice      = "choice"
	promptMultiChoice = "multi-choice"
)

var nilValidator = func(val string) error {
	return nil
}

var errNotInt = errors.New("value is not a valid integer")
var errIntOutOfRange = errors.New("integer is too long")

type Prompt interface {
	kind() string
	Help() string
	Prompt(prompter Prompter) (interface{}, error)
}

type PromptMap map[string]Prompt

type Base struct {
	Msg      string
	HelpText string
}

func (p *Base) Help() string {
	return p.HelpText
}

type input struct {
	Base
}

func (p *input) kind() string {
	return promptInput
}

func (p *input) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.HelpText, nilValidator)
}

type integer struct {
	Base
}

func (p *integer) kind() string {
	return promptInteger
}

func (p *integer) Prompt(prompter Prompter) (interface{}, error) {
	var value int
	_, err := prompter.Get(p.Msg, p.HelpText, func(val string) error {
		if len(val) == 0 {
			return errNotInt
		}
		var err error
		switch value, err = strconv.Atoi(val); {
		case errors.Is(err, strconv.ErrSyntax):
			return errNotInt
		case errors.Is(err, strconv.ErrRange):
			return errIntOutOfRange
		case err != nil:
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

type confirm struct {
	Base
}

func (p *confirm) kind() string {
	return promptConfirm
}

func (p *confirm) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.Confirm(p.Msg, p.HelpText)
}

type choice struct {
	Base
	options []string
}

func (p *choice) kind() string {
	return promptChoice
}

func (p *choice) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectOne(p.Msg, p.HelpText, p.options)
}

type multiChoice struct {
	Base
	options []string
}

func (p *multiChoice) kind() string {
	return promptMultiChoice
}

func (p *multiChoice) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectMultiple(p.Msg, p.HelpText, p.options)
}
