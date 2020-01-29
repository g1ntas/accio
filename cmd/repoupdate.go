package cmd

import (
	"errors"
	"github.com/g1ntas/accio/generator"

	"github.com/spf13/cobra"
)

// updateCmd represents `accio repo update` command
var updateCmd = &cobra.Command{
	Use:   "update [repository]",
	Short: "Update repository with recent changes",
	Long: ``,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo *generator.Repository
		if len(args) > 0 {
			repo, _ = env.registry.Repos[args[0]]
		}
		if repo == nil {
			var err error
			repo, err = promptSelectRepo()
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("all", false, "Update all repositories")
	updateCmd.Flags().String("replace", "", "Replace specified repository with a new path/url")
	updateCmd.Flags().BoolP("help", "h", false, "Show help")
	repoCmd.AddCommand(updateCmd)
}

func promptSelectRepo() (*generator.Repository, error) {
	if len(env.registry.Repos) == 0 {
		return nil, errors.New("no repositories added yet")
	}
	choices, i := make([]string, len(env.registry.Repos)), 0
	for _, r := range env.registry.Repos {
		choices[i] = r.Path
		i++
	}
	r, err := env.prompter.SelectOne("Choose repository:", "", choices)
	if err != nil {
		return nil, err
	}
	repo, _ := env.registry.Repos[r]
	return repo, nil
}
