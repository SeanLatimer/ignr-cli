package cli

import (
	"fmt"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
	"go.seanlatimer.dev/ignr/internal/cache"
	"go.seanlatimer.dev/ignr/internal/templates"
)

func newSearchCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <pattern>",
		Short: "Search templates by name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cachePath, err := cache.InitializeCache()
			if err != nil {
				return err
			}

			items, err := templates.DiscoverTemplates(cachePath)
			if err != nil {
				return err
			}

			pattern := strings.Join(args, " ")
			names := make([]string, 0, len(items))
			for _, item := range items {
				names = append(names, item.Name)
			}

			matches := fuzzy.FindFrom(pattern, stringSource(names))
			for _, match := range matches {
				item := items[match.Index]
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", item.Category, item.Name)
			}
			return nil
		},
	}

	return cmd
}

type stringSource []string

func (s stringSource) Len() int {
	return len(s)
}

func (s stringSource) String(i int) string {
	return s[i]
}
