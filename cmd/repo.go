package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// repoCmd represents `accio repo` command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "A brief description of your command",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("repo called")
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
