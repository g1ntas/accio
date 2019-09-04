package cmd

import (
	"errors"
	"fmt"
	"github.com/g1ntas/accio/generator"
	"os"

	"github.com/spf13/cobra"
)

// addCmd represents `accio repo add` command
var addCmd = &cobra.Command{
	Use:   "add {path_or_url}",
	Short: "Add a new global repository and cache it",
	Long: ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 0 {
			return errors.New("directory path must be specified as a first argument")
		}
		info, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("specified path must be a directory, not a file")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// todo: add cosmetics: better messages, more information, maybe colors
		repo := generator.NewFileSystemRepository(args[0])
		if err := repo.Parse(); err != nil {
			return err
		}
		registry.AddRepository(repo)
		err := registry.Save()
		if err != nil {
			return err
		}
		fmt.Println("Done.")

		return nil
	},
}

func init() {
	repoCmd.AddCommand(addCmd)
}
