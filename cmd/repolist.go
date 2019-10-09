package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents `accio repo list` command
var listRepositoriesCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all repositories",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(registry.Repos) <= 0 {
			fmt.Println("No repositories found")
			return
		}
		fmt.Println("Repositories:")
		i := 1
		for _, r := range registry.Repos {
			fmt.Printf("%d. %s\n", i, r.Path)
			i++
		}
	},
}

func init() {
	repoCmd.AddCommand(listRepositoriesCmd)
}
