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
	promptList        = "list"
	promptChoice      = "choice"
	promptMultiChoice = "multi-choice"
)

var nilValidator = func(val string) error {
	return nil
}

type Prompt interface {
	kind() string
	Prompt(prompter Prompter) (interface{}, error)
}

type PromptMap map[string]Prompt

type base struct {
	msg  string
	help string
}

// input
type input struct {
	base
}

func (p *input) kind() string {
	return promptInput
}

func (p *input) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.msg, p.help, nilValidator)
}


// integer
type integer struct {
	base
}

func (p *integer) kind() string {
	return promptInteger
}

func (p *integer) Prompt(prompter Prompter) (interface{}, error) {
	val, err := prompter.Get(p.msg, p.help, func(val string) error {
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
	base
}

func (p *confirm) kind() string {
	return promptConfirm
}

func (p *confirm) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.Confirm(p.msg, p.help)
}


// list
type list struct {
	base
}

func (p *list) kind() string {
	return promptList
}

func (p *list) Prompt(prompter Prompter) (interface{}, error) {
	val, err := prompter.Get(p.msg, p.help, nilValidator)
	// todo: split string by comma
	return val, err
}


// choice
type choice struct {
	base
	options []string
}

func (p *choice) kind() string {
	return promptChoice
}

func (p *choice) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectOne(p.msg, p.help, p.options)
}


// multiChoice
type multiChoice struct {
	base
	options []string
}

func (p *multiChoice) kind() string {
	return promptMultiChoice
}

func (p *multiChoice) Prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectMultiple(p.msg, p.help, p.options)
}