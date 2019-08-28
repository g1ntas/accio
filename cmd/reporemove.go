package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// removeCmd represents `accio repo remove` command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove repositories by given index(-es). Prompts for confirmation. ",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("remove called")
	},
}

func init() {
	repoCmd.AddCommand(removeCmd)

	listCmd.Flags().BoolP("all", "", false, "Remove all repositories")
	listCmd.Flags().BoolP("force", "f", false, "Force remove - doesn't ask for confirmation")
}
