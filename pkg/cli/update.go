package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.seanlatimer.dev/ignr/internal/cache"
)

func newUpdateCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the cached gitignore templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cachePath, err := cache.UpdateCache()
			if err != nil {
				return err
			}
			status, err := cache.GetStatus()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated cache at %s\n", cachePath)
			if status.HeadCommit != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "HEAD %s\n", status.HeadCommit)
			}
			return nil
		},
	}

	return cmd
}
