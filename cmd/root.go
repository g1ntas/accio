package cmd

import (
	"fmt"
	"github.com/g1ntas/accio/generators"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var registry *generators.Registry

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "accio",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var err error
	registry, err = loadRegistry()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err = rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadRegistry() (*generators.Registry, error) {
	path, err := registryPath()
	if err != nil {
		return nil, err
	}
	reg := generators.NewRegistry(path)
	if err = reg.Load(); err != nil {
		return nil, err
	}
	return reg, nil
}

// userConfigDir returns base directory for storing user configuration for accio.
func userConfigDir() (string, error) {
	// todo: make directory configurable, and use this func value as default
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "accio"), nil
}

// registryPath returns file path to the config file containing information about existing generators and repositories.
func registryPath() (string, error) {
	dir, err := userConfigDir()
	if err != nil {
		return "", err
	}
	// todo: move filename or extension in generators registry?
	return filepath.Join(dir, "registry.json"), nil
}
