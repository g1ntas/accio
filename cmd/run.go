package cmd

import (
	"errors"
	"fmt"
	"github.com/g1ntas/accio/fs"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/generator/model"
	"github.com/g1ntas/accio/prompter"
	"github.com/spf13/cobra"
	"os"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run {generators}",
	Short: "Run an existing generators from one of the repositories",
	Long: ``,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// todo: apply working dir flag
		writeDir, err := os.Getwd()
		if err != nil {
			return err
		}
		promp := prompter.NewCLIPrompter()
		generators := registry.FindGenerators(args[0])
		if len(generators) == 0 {
			return errors.New("no generators found with a given name")
		}
		var gen *generator.Generator

		if len(generators) > 1 {
			fmt.Println("Multiple generators found:")

			choices := make([]string, len(generators))
			for i, g := range generators {
				choices[i] = fmt.Sprintf("%s: %s", g.Name, g.Description)
			}
			i, err := promp.SelectOneIndex("Choose one:", "", choices)
			if err != nil {
				return err
			}
			gen = generators[i]
		} else {
			gen = generators[0]
		}
		data, err := gen.PromptAll(promp)
		if err != nil {
			return err
		}
		parser, err := model.NewParser(data)
		if err != nil {
			return err
		}
		filesystem := fs.NewNativeFS()
		runner := generator.NewRunner(filesystem, parser, writeDir, func(path string) bool {
			fmt.Printf("File at path %q already exists\n", path)
			overwrite, err := promp.Confirm("Do you want to overwrite it?", "")
			if err != nil {
				fmt.Print(err)
				return false
			}
			return overwrite
		})

		err = runner.Run(gen)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolP("force", "f", false, "Overwrite existing paths without asking confirmation")

	// todo: get current directory and set as default flag value
	runCmd.Flags().StringP("working-dir", "w", "", "Specify working directory")
}
