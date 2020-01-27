package generator

import (
	"errors"
	"fmt"
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
	promptList        = "list"
	promptChoice      = "choice"
	promptMultiChoice = "multi-choice"
)

var nilValidator = func(val string) error {
	return nil
}

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

// input
type input struct {
	Base
}

func (p *input) kind() string {
	return promptInput
}

func (p *input) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.HelpText, nilValidator)
}


// integer
type integer struct {
	Base
}

func (p *integer) kind() string {
	return promptInteger
}

func (p *integer) Prompt(prompter Prompter) (interface{}, error) {
	val, err := prompter.Get(p.Msg, p.HelpText, func(val string) error {
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


// confirm
type confirm struct {
	Base
}

func (p *confirm) kind() string {
	return promptConfirm
}

func (p *confirm) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.Confirm(p.Msg, p.HelpText)
}


// list
type list struct {
	Base
}

func (p *list) kind() string {
	return promptList
}

func (p *list) Prompt(prompter Prompter) (interface{}, error) {
	values := make([]string, 0, 5)
	help := fmt.Sprintf("Return empty value to finish.\n\n%s", p.HelpText)
	for {
		val, err := prompter.Get(p.Msg, help, nilValidator)
		if err != nil {
			return nil, err
		}
		if val == "" {
			return values, nil
		}
		values = append(values, val)
	}
}

// choice
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


// multiChoice
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