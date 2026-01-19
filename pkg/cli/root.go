package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	ConfigPath string
	Verbose    bool
	Quiet      bool
}

var Version = "dev"

func Execute() error {
	opts := &Options{}
	root := NewRootCommand(opts)
	return root.Execute()
}

func NewRootCommand(opts *Options) *cobra.Command {
	root := &cobra.Command{
		Use:   "ignr",
		Short: "Offline-first gitignore generator",
	}

	root.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "Config file path")
	root.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	root.PersistentFlags().BoolVar(&opts.Quiet, "quiet", false, "Suppress non-error output")

	root.AddCommand(
		newListCommand(opts),
		newSearchCommand(opts),
		newGenerateCommand(opts),
		newPresetCommand(opts),
		newUpdateCommand(opts),
	)

	root.Version = Version
	root.SetVersionTemplate(fmt.Sprintf("ignr %s\n", Version))

	return root
}

func ExitWithError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
