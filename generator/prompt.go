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
	promptFile        = "file"
)

var nilValidator = func(val string) error {
	return nil
}

type prompt interface {
	kind() string
	prompt(prompter Prompter) (interface{}, error)
}

type PromptMap map[string]prompt

type Base struct {
	Msg  string `toml:"message"`
	Help string `toml:"help"`
}

// Input
type Input struct {
	Base
}

func (p *Input) kind() string {
	return promptInput
}

func (p *Input) prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.Help, nilValidator)
}


// Integer
type Integer struct {
	Base
}

func (p *Integer) kind() string {
	return promptInteger
}

func (p *Integer) prompt(prompter Prompter) (interface{}, error) {
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
	Base
}

func (p *Confirm) kind() string {
	return promptConfirm
}

func (p *Confirm) prompt(prompter Prompter) (interface{}, error) {
	return prompter.Confirm(p.Msg, p.Help)
}


// List
type List struct {
	Base
}

func (p *List) kind() string {
	return promptList
}

func (p *List) prompt(prompter Prompter) (interface{}, error) {
	val, err := prompter.Get(p.Msg, p.Help, nilValidator)
	// todo: split string by comma
	return val, err
}


// Choice
type Choice struct {
	Base
	Options []string `toml:"options"`
}

func (p *Choice) kind() string {
	return promptChoice
}

func (p *Choice) prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectOne(p.Msg, p.Help, p.Options)
}


// MultiChoice
type MultiChoice struct {
	Base
	Options []string `toml:"options"`
}

func (p *MultiChoice) kind() string {
	return promptMultiChoice
}

func (p *MultiChoice) prompt(prompter Prompter) (interface{}, error) {
	return prompter.SelectMultiple(p.Msg, p.Help, p.Options)
}


// File
type File struct {
	Base
}

func (p *File) kind() string {
	return promptFile
}

func (p *File) prompt(prompter Prompter) (interface{}, error) {
	return prompter.Get(p.Msg, p.Help, func(val string) error {
		// todo: validate is an existing and readable file
		return nil
	})
}

func (m PromptMap) UnmarshalTOML(data interface{}) error {
	prompts := data.(map[string]interface{})
	for key, def := range prompts {
		mapping := def.(map[string]interface{})
		typ, ok := mapping["type"].(string)
		if !ok {
			return fmt.Errorf("prompt %q has no type", key)
		}
		base := Base{}
		base.Msg, _ = mapping["message"].(string)
		base.Help, _ = mapping["help"].(string)
		switch typ {
		case promptInput:
			m[key] = &Input{base}
		case promptInteger:
			m[key] = &Integer{base}
		case promptConfirm:
			m[key] = &Confirm{base}
		case promptList:
			m[key] = &List{base}
		case promptChoice:
			options, _ := mapping["options"].([]string)
			m[key] = &Choice{base, options}
		case promptMultiChoice:
			options, _ := mapping["options"].([]string)
			m[key] = &MultiChoice{base, options}
		case promptFile:
			m[key] = &File{base}
		default:
			return fmt.Errorf("prompt %q with unknown type %q", key, typ)
		}
	}
	return nil
}