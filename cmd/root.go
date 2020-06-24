package cmd

import (
	"fmt"
	"github.com/g1ntas/accio/prompter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
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
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// In addition, it handles errors returned by inner commands by printing them.
func Execute() {
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
