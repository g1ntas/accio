package cmd

import (
	"github.com/g1ntas/accio/fs"
	"github.com/g1ntas/accio/prompter"
	"github.com/spf13/cobra"
	"os"
)

var env environment

type environment struct {
	fs fs.Filesystem
	prompter *prompter.CLI
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "accio",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	env.fs = fs.NewNativeFS()
	env.prompter = prompter.NewCLIPrompter()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}