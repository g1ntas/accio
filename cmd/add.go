package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents `accio repo add` command
var addCmd = &cobra.Command{
	Use:   "add {path_or_url}",
	Short: "Add a new global repository and cache it",
	Long: ``,
	Args: func(cmd *cobra.Command, args []string) error {
		// todo: validate repository is valid:
		// todo: 1. local directory
		// todo: 2. git repository
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// todo: add repo
		fmt.Println("add called")
	},
}

func init() {
	repoCmd.AddCommand(addCmd)
}
