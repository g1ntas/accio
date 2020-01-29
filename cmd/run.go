package cmd

import (
	"fmt"
	"github.com/g1ntas/accio/fs"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/generator/model"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [generator]",
	Short: "Run a generator from given url",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := newGeneratorFromUrl(args[0])
		if err != nil {
			return err
		}
		writeDir, err := getWorkingDir(cmd)
		if err != nil {
			return err
		}
		data, err := gen.PromptAll(env.prompter)
		if err != nil {
			return err
		}
		parser, err := model.NewParser(data)
		if err != nil {
			return err
		}
		runner := generator.NewRunner(
			getFilesystem(cmd),
			parser,
			writeDir,
			getOverwriteHandler(cmd),
		)
		err = runner.Run(gen)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	runCmd.SetHelpFunc(generatorHelpFunc)
	runCmd.Flags().Bool("dry", false, "Run without writing to filesystem")
	runCmd.Flags().BoolP("force", "f", false, "Overwrite existing paths without asking confirmation")
	runCmd.Flags().BoolP("help", "h", false, "Show help")
	runCmd.Flags().StringP("working-dir", "w", "", "Specify working directory")
	rootCmd.AddCommand(runCmd)
}

func generatorHelpFunc(cmd *cobra.Command, args []string) {
	if err := cmd.ValidateArgs(cmd.Flags().Args()); err != nil {
		cmd.Root().HelpFunc()(cmd, args)
		return
	}
	gen, err := newGeneratorFromUrl(args[1])
	if err != nil {
		cmd.PrintErrln(err) // todo: refactor to use own print fn
		if err := cmd.Usage(); err != nil {
			cmd.PrintErrln(err) // todo: refactor to use own print fn
		}
		return
	}
	helpCmd := &cobra.Command{
		Use:   gen.Name,
		Short: gen.Description,
		Long:  buildGeneratorHelp(gen),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(helpCmd)
	helpCmd.Flags().AddFlagSet(cmd.Flags())
	helpCmd.SetHelpFunc(cmd.Root().HelpFunc())
	helpCmd.HelpFunc()(helpCmd, args)
}

func getFilesystem(cmd *cobra.Command) fs.Filesystem {
	if cmd.Flag("dry").Value.String() == "true" {
		return fs.NewDryFS(env.fs)
	}
	return env.fs
}

func getOverwriteHandler(cmd *cobra.Command) generator.OverwriteFn {
	force := cmd.Flag("force").Value.String() == "true"
	return func(path string) bool {
		if force {
			return true
		}
		fmt.Printf("File at path %q already exists\n", path)
		overwrite, err := env.prompter.Confirm("Do you want to overwrite it?", "")
		if err != nil {
			fmt.Print(err)
			return false
		}
		return overwrite
	}
}

func getWorkingDir(cmd *cobra.Command) (string, error) {
	dir := cmd.Flag("working-dir").Value.String()
	if dir != "" {
		return dir, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func buildGeneratorHelp(gen *generator.Generator) string {
	help := strings.TrimSpace(gen.Help)
	if len(gen.Prompts) == 0 {
		return help
	}
	help += "\n\nPrompts:\n"
	for name, pr := range gen.Prompts {
		help += fmt.Sprintf("[%s]\n", name)
		switch h := pr.Help(); {
		case len(h) > 0:
			help += h
		default:
			help += "This prompt has no description"
		}
		help += "\n\n"
	}
	return strings.TrimSpace(help)
}

func newGeneratorFromUrl(url string) (*generator.Generator, error) {
	// todo: create generator by given URL
	return &generator.Generator{}, nil
}