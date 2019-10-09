package cmd

import (
	"github.com/spf13/cobra"
)

// removeCmd represents `accio repo remove` command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove repositories by given index(-es) or origin. Prompts for confirmation. ",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		//cmd.Flag()
	},
}

func init() {
	repoCmd.AddCommand(removeCmd)

	listCmd.Flags().BoolP("force", "f", false, "Force remove - doesn't ask for confirmation")
}
