package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.seanlatimer.dev/ignr/internal/cache"
	"go.seanlatimer.dev/ignr/internal/templates"
)

func newListCommand(opts *Options) *cobra.Command {
	var category string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available gitignore templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cachePath, err := cache.InitializeCache()
			if err != nil {
				return err
			}

			items, err := templates.DiscoverTemplates(cachePath)
			if err != nil {
				return err
			}

			categoryFilter := strings.ToLower(strings.TrimSpace(category))
			for _, item := range items {
				if categoryFilter != "" && strings.ToLower(string(item.Category)) != categoryFilter {
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", item.Category, item.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category (root, Global, community)")
	return cmd
}
