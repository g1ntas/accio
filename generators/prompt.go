package generators

import (
	"errors"
	"strconv"
)

type Prompter interface {
	Get(message, help string, validator InputValidatorFunc) (string, error)
	SelectOne(message, help string, options []string) (string, error)
	SelectMultiple(message, help string, options []string) ([]string, error)
	Confirm(message, help string) (bool, error)
}

type InputValidatorFunc func(val string) error

const (
	promptInput       = "input"
	promptInteger     = "integer"
	promptConfirm     = "confirm"
	promptList        = "list"
	promptChoice      = "choice"
	promptMultiChoice = "multi-choice"
	promptFile        = "file"
)

var nilValidator = func(val string) error {
	return nil
}

type prompt interface {
	kind() string
	prompt(prompter Prompter) (interface{}, error)
}

type base struct {
	Msg  string `json:"message",yaml:"message"`
	Help string `json:"help",yaml:"help"`
}

// input
type input struct {
	base
}

func (p *input) kind() string {
	return promptInput
}

func (p *input) prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.Help, nilValidator)
}


// integer
type integer struct {
	base
}

func (p *integer) kind() string {
	return promptInteger
}

func (p *integer) prompt(prompter Prompter) (interface{}, error) {
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


// confirm
type confirm struct {
	base
}

func (p *confirm) kind() string {
	return promptConfirm
}

func (p *confirm) prompt(prompter Prompter) (interface{}, error) {
	return prompter.Confirm(p.Msg, p.Help)
}


// List
type list struct {
	Msg  string `json:"message",yaml:"message"`
	Help string `json:"help",yaml:"help"`
}

func (p *list) kind() string {
	return promptList
}

func (p *list) prompt(prompter Prompter) (interface{}, error) {
	val, err := prompter.Get(p.Msg, p.Help, nilValidator)
	// todo: split string by comma
	return val, err
}


// choice
type choice struct {
	base
	Options []string `json:"options",yaml:"options"`
}

func (p *choice) kind() string {
	return promptChoice
}

func (p *choice) prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectOne(p.Msg, p.Help, p.Options)
}


// multiChoice
type multiChoice struct {
	base
	Options []string `json:"options",yaml:"options"`
}

func (p *multiChoice) kind() string {
	return promptMultiChoice
}

func (p *multiChoice) prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectMultiple(p.Msg, p.Help, p.Options)
}


// file
type file struct {
	base
}

func (p *file) kind() string {
	return promptFile
}

func (p *file) prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.Help, func(val string) error {
		// todo: validate is an existing and readable file
		return nil
	})
}