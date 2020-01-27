package cmd

import (
	"errors"
	"fmt"
	"github.com/g1ntas/accio/fs"
	"github.com/g1ntas/accio/generator"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// addCmd represents `accio repo add` command
var addCmd = &cobra.Command{
	Use:   "add {path_or_url}",
	Short: "Add a new global repository and cache it",
	Long: ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) <= 0 {
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
		filesystem := fs.NewNativeFS() // todo: share between commands

		// todo: add cosmetics: better messages, more information, colors
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		repo := generator.NewRepository(filepath.ToSlash(path))
		c, err := repo.ImportGenerators(filesystem)
		if err != nil {
			return err
		}
		if c <= 0 {
			return fmt.Errorf("no configured generators found in %s\n", path)
		}
		if err := env.registry.AddRepository(repo); err != nil {
			return err
		}
		if err := saveRegistry(env.registry); err != nil {
			return err
		}
		fmt.Printf("Added %d generator(-s)\nDone.\n", c)
		return nil
	},
}

func init() {
	repoCmd.AddCommand(addCmd)
}
