package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents `accio repo list` command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all added repositories with corresponding indexes",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// todo: retrieve
		fmt.Println("list called")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
