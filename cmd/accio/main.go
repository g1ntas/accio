package main

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"

	"github.com/g1ntas/accio/prompter"
)

// env holds dependencies and settings that represents environment for app implementation.
var env environment

type environment struct {
	fs       afero.Afero
	prompter *prompter.CLI
}

// rootCmd represents the base command, which can be executed by running executable without any arguments.
var rootCmd = &cobra.Command{
	Use:   "accio",
	Short: "Accio is a flexible framework for boilerplate code generators.",
	Long: `Accio is a flexible framework for boilerplate code generators.
It is designed with readability in mind because logic-full 
templates are hard to maintain. Its modular approach to 
templates makes it easy and fun to work with, and the possibility 
to script - just powerful enough to handle most edge-cases.

For documentation on how to create new generators, go to the 
official project repository at https://github.com/g1ntas/accio.

`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// main adds all child commands to the root command and sets flags appropriately.
// In addition, it handles errors returned by inner commands by printing them.
func main() {
	env.fs = afero.Afero{Fs: afero.NewOsFs()}
	env.prompter = prompter.NewCLIPrompter(os.Stdin, os.Stdout, os.Stderr)
	if cmd, err := rootCmd.ExecuteC(); err != nil {
		if cmd == nil {
			printErr(err)
			os.Exit(1)
		}
		printErr(err)
		fmt.Println(cmd.UsageString())
		os.Exit(1)
	}
}

func printErr(e error) {
	format := "[ERROR] %s\n"
	_, err := fmt.Fprintf(os.Stderr, format, e.Error())
	if err != nil {
		fmt.Printf(format, e.Error())
	}
}
