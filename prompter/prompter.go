package prompter

import (
	"github.com/AlecAivazis/survey"
)

type CLI struct {
}

func NewCLIPrompter() *CLI {
	return &CLI{}
}

func setDefaultStyle(icons *survey.IconSet) {
	icons.Error.Text = "[ERROR]"
	icons.Help.Text = ""
	icons.Question.Text = "?"
	icons.SelectFocus.Text = ">"
	icons.UnmarkedOption.Text = "[]"
	icons.UnmarkedOption.Text = "[x]"

	icons.Error.Format = ""
	icons.Help.Format = ""
	icons.HelpInput.Format = ""
	icons.Question.Format = ""
	icons.SelectFocus.Format = ""
	icons.UnmarkedOption.Format = ""
	icons.UnmarkedOption.Format = ""
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
	}), survey.WithIcons(setDefaultStyle))
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
	err = survey.AskOne(prompt, &val, survey.WithIcons(setDefaultStyle))
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
	err = survey.AskOne(prompt, &val, survey.WithIcons(setDefaultStyle))
	if err != nil {
		return "", err
	}
	return
}

func (p *CLI) SelectOneIndex(message, help string, options []string) (val int, err error) {
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err = survey.AskOne(prompt, &val, survey.WithIcons(setDefaultStyle))
	if err != nil {
		return -1, err
	}
	return
}

func (p *CLI) SelectMultiple(message, help string, options []string) ([]string, error) {
	val := make([]string, len(options), 0)
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &val, survey.WithIcons(setDefaultStyle))
	if err != nil {
		return []string{}, err
	}
	return val, nil
}

