package cmd

import (
	"fmt"
	"github.com/g1ntas/accio/generators"
	"github.com/g1ntas/accio/generators/gob"
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
	return filepath.Join(dir, "registry.gob"), nil
}

// loadRegistry reads registry file from user config directory and stores data in struct.
// If registry file does not exist within user config directory,
// then new registry struct is returned instead.
func loadRegistry() (*generators.Registry, error) {
	path, err := registryPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return generators.NewRegistry(), nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	reg, err := gob.Unserialize(f)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// Save stores all the data from registry struct into configured registry file.
func saveRegistry(reg *generators.Registry) error {
	path, err := registryPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	err = gob.Serialize(f, reg)
	if err != nil {
		return err
	}
	return nil
}
