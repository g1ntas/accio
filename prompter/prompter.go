package prompter

import (
	"github.com/AlecAivazis/survey"
)

type CLI struct {
}

func NewCLIPrompter() *CLI {
	return &CLI{}
}

func (p *CLI) Get(message, help string, validate func(val string) error) (val string, err error) {
	prompt := &survey.Input{
		Message: message,
		Help: help,
	}
	err = survey.AskOne(prompt, &val, survey.WithValidator(func(answer interface{}) error {
		if err := validate(answer.(string)); err != nil {
			return err
		}
		return nil
	}))
	if err != nil {
		return "", err
	}
	return
}

func (p *CLI) Confirm(message, help string) (val bool, err error) {
	prompt := &survey.Confirm{
		Message: message,
		Help: help,
	}
	err = survey.AskOne(prompt, &val)
	if err != nil {
		return false, err
	}
	return
}

func (p *CLI) SelectOne(message, help string, options []string) (val string, err error) {
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err = survey.AskOne(prompt, &val)
	if err != nil {
		return "", err
	}
	return
}

func (p *CLI) SelectMultiple(message, help string, options []string) ([]string, error) {
	val := make([]string, len(options), 0)
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &val)
	if err != nil {
		return []string{}, err
	}
	return val, nil
}

