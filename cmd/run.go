package cmd

import (
	"errors"
	"fmt"
	"github.com/g1ntas/accio/fs"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/prompter"
	"github.com/g1ntas/accio/template"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run {generators}",
	Short: "Run an existing generators from one of the repositories",
	Long: ``,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filesystem := fs.NewNativeFS()
		promp := prompter.NewCLIPrompter()
		tpl := template.Engine{}

		generators := registry.FindGenerators(args[0])
		if len(generators) == 0 {
			return errors.New("no generators found with a given name")
		}
		var gen *generator.Generator
		if len(generators) > 1 {
			// todo: prompt to choose one
		} else {
			gen = generators[0]
		}
		runner := generator.NewRunner(promp, filesystem, tpl)
		// todo: run generators
		fmt.Println("run called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolP("force", "f", false, "Overwrite existing paths without asking confirmation")

	// todo: get current directory and set as default flag value
	runCmd.Flags().StringP("working-dir", "w", "", "Specify working directory")
}
