package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-safetemp"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/generator/blueprint"
	"github.com/g1ntas/accio/gitgetter"
)

var runCmd = &cobra.Command{
	Use:   "run [generator]",
	Short: "Run a generator from directory or git repository",
	Long: `Executes generator writing all generated files at the current 
directory, unless specified otherwise. If the generator has 
any prompts configured, then they will be prompted first.

Command accepts a single required argument specifying the 
location of the generator. Location can be either a path 
to a local directory or an URL to a git repository.

For git repository URLs:
  Conditions:
  - Git client is required, which must be registered in a 
    PATH environment variable and be accessible as 'git' 
    executable;
  - HTTP, HTTPS, and SSH protocols are supported. 
    SSH protocol is only limited to username 'git', 
    e.g. 'git@github.com'.

  Subdirectories are supported and can be specified after 
  double-slash '//'. In the case of 'github.com' or 
  'bitbucket.org', only a single slash '/' can be used.
  
  Example:
  github.com/g1ntas/accio/examples/open-source-license
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, closer, err := fetchGeneratorFromUrl(args[0])
		if err != nil {
			return err
		}
		defer closeGracefully(closer)
		err = gen.ReadConfig(env.fs)
		if err != nil {
			return err
		}
		writeDir, err := workingDir(cmd)
		if err != nil {
			return err
		}
		data, err := gen.PromptAll(env.prompter)
		if err != nil {
			return err
		}
		parser, err := blueprint.NewParser(data)
		if err != nil {
			return err
		}
		runner := generator.NewRunner(
			filesystem(cmd),
			parser,
			writeDir,
			generator.OnFileExists(existingFileHandler(cmd)),
			generator.OnError(errorHandler(cmd)),
			generator.OnSuccess(successHandler(cmd)),
			generator.IgnoreDir(".git"),
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
	runCmd.Flags().BoolP("ignore-errors", "i", false, "Ignore errors for files being generated")
	runCmd.Flags().StringP("working-dir", "w", "", "Specify working directory")
	rootCmd.AddCommand(runCmd)
}

func fetchGeneratorFromUrl(src string) (*generator.Generator, io.Closer, error) {
	info, err := env.fs.Stat(src)
	if err == nil && info.IsDir() {
		return generator.NewGenerator(src), ioutil.NopCloser(nil), nil
	}
	dst, closer, err := safetemp.Dir("", "accio")
	if err != nil {
		return nil, nil, err
	}
	dst = filepath.Join(dst, "tmp") // work around for https://github.com/hashicorp/go-getter/issues/114
	err = cloneRepo(src, dst)
	if err != nil {
		closeGracefully(closer)
		return nil, nil, err
	}
	gen := generator.NewGenerator(dst)
	return gen, closer, nil
}

func closeGracefully(c io.Closer) {
	if err := c.Close(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}

// generatorHelpFunc defines run command's help behaviour.
// When --help flag is provided together with at least single argument,
// generator will be parsed and help text from configuration file will be shown.
// In case of no arguments, default help behaviour will be executed.
func generatorHelpFunc(cmd *cobra.Command, args []string) {
	if err := cmd.ValidateArgs(cmd.Flags().Args()); err != nil {
		cmd.Root().HelpFunc()(cmd, args)
		return
	}
	gen, closer, err := fetchGeneratorFromUrl(args[1])
	if err != nil {
		printErr(err)
		fmt.Println(cmd.UsageString())
		return
	}
	defer closeGracefully(closer)
	err = gen.ReadConfig(env.fs)
	if err != nil {
		printErr(err)
		return
	}
	helpCmd := &cobra.Command{
		Use:   args[1],
		Short: "",
		Long:  buildGeneratorHelp(gen),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(helpCmd)
	helpCmd.Flags().AddFlagSet(cmd.Flags())
	helpCmd.SetHelpFunc(cmd.Root().HelpFunc())
	helpCmd.HelpFunc()(helpCmd, args)
}

func filesystem(cmd *cobra.Command) afero.Afero {
	if getBoolFlag(cmd, "dry") {
		roBase := afero.NewReadOnlyFs(env.fs.Fs)
		ufs := afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
		return afero.Afero{Fs: ufs}
	}
	return env.fs
}

func existingFileHandler(cmd *cobra.Command) generator.OnExistsFn {
	force := getBoolFlag(cmd, "force")
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

func errorHandler(cmd *cobra.Command) generator.OnErrorFn {
	ignoreErr := getBoolFlag(cmd, "ignore-errors")
	return func(err error) bool {
		if ignoreErr {
			printErr(fmt.Errorf("%w. Skipping...", err))
		}
		return ignoreErr
	}
}

func successHandler(_ *cobra.Command) generator.OnSuccessFn {
	return func(_, dst string) {
		fmt.Printf("[SUCCESS] %s created.\n", dst)
	}
}

func workingDir(cmd *cobra.Command) (string, error) {
	dir := getStringFlag(cmd, "working-dir")
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

func cloneRepo(src, dst string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	client := &gitgetter.Client{Pwd: cwd}
	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.CloneRepository(ctx, src, dst); err != nil {
			errChan <- err
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		return errors.New(sig.String())
	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		return err
	}
	return nil
}
