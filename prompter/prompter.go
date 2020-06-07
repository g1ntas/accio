package prompter

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"io"
)

type CLI struct {
	Stdin  terminal.FileReader
	Stdout terminal.FileWriter
	Stderr io.Writer
}

func NewCLIPrompter(in terminal.FileReader, out terminal.FileWriter, err io.Writer) *CLI {
	return &CLI{Stdin: in, Stdout: out, Stderr: err}
}

func setDefaultStyle(icons *survey.IconSet) {
	icons.Error.Text = "[ERROR]"
	icons.Error.Format = ""
}

func (p *CLI) Get(message, help string, validate func(val string) error) (val string, err error) {
	prompt := &survey.Input{
		Message: message,
		Help:    help,
	}
	validator := func(answer interface{}) error {
		if err := validate(answer.(string)); err != nil {
			return err
		}
		return nil
	}
	err = survey.AskOne(prompt, &val,
		survey.WithValidator(validator),
		survey.WithIcons(setDefaultStyle),
		survey.WithStdio(p.Stdin, p.Stdout, p.Stderr),
	)
	if err != nil {
		return "", err
	}
	return
}

func (p *CLI) Confirm(message, help string) (val bool, err error) {
	prompt := &survey.Confirm{
		Message: message,
		Help:    help,
	}
	err = survey.AskOne(prompt, &val,
		survey.WithIcons(setDefaultStyle),
		survey.WithStdio(p.Stdin, p.Stdout, p.Stderr),
	)
	if err != nil {
		return false, err
	}
	return
}

func (p *CLI) SelectOne(message, help string, options []string) (val string, err error) {
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Help:    help,
	}
	err = survey.AskOne(prompt, &val,
		survey.WithIcons(setDefaultStyle),
		survey.WithStdio(p.Stdin, p.Stdout, p.Stderr),
	)
	if err != nil {
		return "", err
	}
	return
}

func (p *CLI) SelectMultiple(message, help string, options []string) ([]string, error) {
	var val []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
		Help:    help,
	}
	err := survey.AskOne(prompt, &val,
		survey.WithIcons(setDefaultStyle),
		survey.WithStdio(p.Stdin, p.Stdout, p.Stderr),
	)
	if err != nil {
		return []string{}, err
	}
	return val, nil
}
