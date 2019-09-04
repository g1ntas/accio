package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run {generators}",
	Short: "Run an existing generators from one of the repositories",
	Long: ``,
	Args: func(cmd *cobra.Command, args []string) error {
		// todo: validate generators exists with a given name
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// todo: lookup for a named generators
		// todo: if multiple generators found prompt to choose one
		// todo: run generators
		fmt.Println("run called")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolP("force", "f", false, "Overwrite existing paths without asking confirmation")

	// todo: get current directory and set as default flag value
	runCmd.Flags().StringP("working-dir", "w", "", "Specify working directory")
}
