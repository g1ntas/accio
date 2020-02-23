package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/generator/blueprint"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-getter/helper/url"
	"github.com/hashicorp/go-safetemp"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [generator]",
	Short: "Run a generator from given url",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dst, closer, err := safetemp.Dir("", "accio")
		if err != nil {
			return err
		}
		dst = filepath.Join(dst, "tmp") // work around for https://github.com/hashicorp/go-getter/issues/114
		defer Close(closer)
		gen, err := fetchGeneratorFromUrl(args[0], dst)
		if err != nil {
			return err
		}
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

	getter.Detectors = []getter.Detector{}
	getter.Getters = map[string]getter.Getter{}
}

func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}

func generatorHelpFunc(cmd *cobra.Command, args []string) {
	if err := cmd.ValidateArgs(cmd.Flags().Args()); err != nil {
		cmd.Root().HelpFunc()(cmd, args)
		return
	}
	gen, err := fetchGeneratorFromUrl(args[1], "")
	if err != nil {
		cmd.PrintErrln(err) // todo: refactor to use own print fn
		if err := cmd.Usage(); err != nil {
			cmd.PrintErrln(err) // todo: refactor to use own print fn
		}
		return
	}

	// todo: remove use and short description
	helpCmd := &cobra.Command{
		Use:   "gen.Name",
		Short: "gen.Description",
		Long:  buildGeneratorHelp(gen),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(helpCmd)
	helpCmd.Flags().AddFlagSet(cmd.Flags())
	helpCmd.SetHelpFunc(cmd.Root().HelpFunc())
	helpCmd.HelpFunc()(helpCmd, args)
}

func filesystem(cmd *cobra.Command) afero.Afero {
	if cmd.Flag("dry").Value.String() == "true" {
		roBase := afero.NewReadOnlyFs(env.fs.Fs)
		ufs := afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
		return afero.Afero{Fs: ufs}
	}
	return env.fs
}

func existingFileHandler(cmd *cobra.Command) generator.OnExistsFn {
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

func errorHandler(cmd *cobra.Command) generator.OnErrorFn {
	ignoreErr := cmd.Flag("ignore-errors").Value.String() == "true"
	return func(err error) bool {
		if ignoreErr {
			printErr(fmt.Errorf("%w. Skipping...", err))
		}
		return ignoreErr
	}
}

func workingDir(cmd *cobra.Command) (string, error) {
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

func fetchGeneratorFromUrl(src, dst string) (*generator.Generator, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	abspath, isFile, err := parseFilePath(src, cwd) // todo: it's possible and target path doesnt't exist and is git
	if err != nil {
		return nil, err
	}
	if isFile {
		if _, err = env.fs.Stat(abspath); err != nil {
			return nil, err
		}
		return generator.NewGenerator(abspath), nil
	}
	gen := generator.NewGenerator(dst)
	ctx, cancel := context.WithCancel(context.Background())
	client := &getter.Client{
		Ctx:     ctx,
		Src:     src,
		Dst:     dst,
		Pwd:     cwd,
		Mode:    getter.ClientModeDir,
		Options: []getter.ClientOption{},
		Detectors: []getter.Detector{
			new(getter.GitHubDetector),
			new(getter.BitBucketDetector),
			new(getter.GitDetector),
		},
		Getters: map[string]getter.Getter{
			"git":   new(getter.GitGetter),
		},
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		return nil, errors.New(sig.String())
	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		return nil, err
	}
	return gen, nil
}

// parseFilePath returns absolute path of the given filepath.
// If filepath is relative path, it will be expanded where
// base path is considered second argument 'pwd'. If given
// path isn't correct filepath, second return parameter
// will be returned as false.
func parseFilePath(p, pwd string) (string, bool, error) {
	d := getter.FileDetector{}
	p, isFile, err := d.Detect(p, pwd)
	if err != nil {
		return "", false, err
	}
	if !isFile {
		return "", false, nil
	}
	u, err := url.Parse(p)
	if err != nil {
		return "", false, err
	}
	return u.Path, true, nil
}
