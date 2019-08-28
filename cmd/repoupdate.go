package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd represents `accio repo update` command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull latest changes from added repositories",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
	},
}

func init() {
	repoCmd.AddCommand(updateCmd)
}
