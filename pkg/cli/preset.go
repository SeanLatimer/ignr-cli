package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.seanlatimer.dev/ignr/internal/cache"
	"go.seanlatimer.dev/ignr/internal/config"
	"go.seanlatimer.dev/ignr/internal/presets"
	"go.seanlatimer.dev/ignr/internal/templates"
	"go.seanlatimer.dev/ignr/internal/tui"
)

func newPresetCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage template presets",
	}

	cmd.AddCommand(
		newPresetCreateCommand(opts),
		newPresetListCommand(opts),
		newPresetShowCommand(opts),
		newPresetDeleteCommand(opts),
		newPresetUseCommand(opts),
	)
	return cmd
}

func newPresetCreateCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name> <template1> [template2...]",
		Short: "Create a preset from template names",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			templateNames := args[1:]

			items, err := discoverAllTemplates()
			if err != nil {
				return err
			}

			index := templates.BuildIndex(items)
			for _, tmpl := range templateNames {
				if _, ok := templates.FindTemplate(index, tmpl); !ok {
					return fmt.Errorf("template not found: %s", tmpl)
				}
			}

			if err := presets.CreatePreset(name, templateNames); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created preset %s with %d templates\n", name, len(templateNames))
			return nil
		},
	}
}

func newPresetListCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := presets.ListPresets()
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No presets found.")
				return nil
			}
			for _, preset := range list {
				fmt.Fprintf(cmd.OutOrStdout(), "%s (%d templates)\n", preset.Name, len(preset.Templates))
			}
			return nil
		},
	}
}

func newPresetShowCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show preset details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			preset, ok, err := presets.FindPreset(name)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("preset not found: %s", name)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", preset.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Templates: %s\n", strings.Join(preset.Templates, ", "))
			if preset.Created != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", preset.Created)
			}
			if preset.Updated != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s\n", preset.Updated)
			}
			return nil
		},
	}
}

func newPresetDeleteCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := presets.DeletePreset(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted preset %s\n", name)
			return nil
		},
	}
}

func newPresetUseCommand(opts *Options) *cobra.Command {
	var output string
	var appendMode bool
	var noHeader bool
	var force bool

	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Generate a .gitignore using a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			preset, ok, err := presets.FindPreset(name)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("preset not found: %s", name)
			}

			items, err := discoverAllTemplates()
			if err != nil {
				return err
			}

			selected, _, err := selectTemplates(preset.Templates, items, nil, nil, true)
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				return fmt.Errorf("no templates selected")
			}

			loaded, err := templates.LoadTemplates(selected)
			if err != nil {
				return err
			}

			target, err := resolveOutputPath(output)
			if err != nil {
				return err
			}

			content := templates.MergeTemplates(loaded, templates.MergeOptions{
				Deduplicate: true,
				AddHeader:   !noHeader,
				Generator:   "ignr",
				Version:     Version,
				Timestamp:   time.Now(),
			})

			if err := handleExistingOutput(cmd, target, appendMode, force, false); err != nil {
				if errors.Is(err, tui.ErrCancelled) {
					return nil
				}
				return err
			}

			if err := writeOutput(target, content, appendMode, force); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated %s with %d templates\n", target, len(selected))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: .gitignore)")
	cmd.Flags().BoolVar(&appendMode, "append", false, "Append to existing file instead of overwrite")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Skip generator header")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file without prompt")
	return cmd
}

func discoverAllTemplates() ([]templates.Template, error) {
	cachePath, err := cache.InitializeCache()
	if err != nil {
		return nil, err
	}

	items, err := templates.DiscoverTemplates(cachePath)
	if err != nil {
		return nil, err
	}

	userPath, err := config.GetUserTemplatePath()
	if err != nil {
		return nil, err
	}
	userItems, err := templates.DiscoverUserTemplates(userPath)
	if err != nil {
		return nil, err
	}

	return append(items, userItems...), nil
}
