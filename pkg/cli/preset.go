package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
	createCmd := newPresetCreateCommand(opts)
	editCmd := newPresetEditCommand(opts)
	listCmd := newPresetListCommand(opts)
	showCmd := newPresetShowCommand(opts)
	deleteCmd := newPresetDeleteCommand(opts)
	useCmd := newPresetUseCommand(opts)

	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage template presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := tui.ShowPresetApp()
			if err != nil {
				if errors.Is(err, tui.ErrCancelled) {
					return nil
				}
				return err
			}
			return nil
		},
	}

	cmd.AddCommand(
		createCmd,
		editCmd,
		listCmd,
		showCmd,
		deleteCmd,
		useCmd,
	)
	return cmd
}

func newPresetCreateCommand(opts *Options) *cobra.Command {
	var noInteractive bool
	cmd := &cobra.Command{
		Use:   "create [name] [template1 template2...]",
		Short: "Create a preset from template names",
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			templateNames := []string{}
			if len(args) > 0 {
				name = args[0]
				if len(args) > 1 {
					templateNames = args[1:]
				}
			}

			items, err := discoverAllTemplates()
			if err != nil {
				return err
			}

			if len(templateNames) > 0 || noInteractive {
				if strings.TrimSpace(name) == "" {
					return fmt.Errorf("preset name is required in non-interactive mode")
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
			}

			existingKeys, err := presetKeys()
			if err != nil {
				return err
			}

			if strings.TrimSpace(name) == "" {
				name, err = tui.ShowPresetNameInput("Preset name:", existingKeys, false)
				if err != nil {
					if errors.Is(err, tui.ErrCancelled) {
						return nil
					}
					return err
				}
			} else {
				key := presets.SluggifyName(name)
				if presetKeyExists(existingKeys, key) {
					return fmt.Errorf("preset key already exists: %s", key)
				}
			}

			selected, err := tui.ShowInteractiveSelector(items, nil, nil, nil)
			if err != nil {
				if errors.Is(err, tui.ErrCancelled) {
					return nil
				}
				return err
			}
			if len(selected) == 0 {
				return fmt.Errorf("no templates selected")
			}

			templateNames = make([]string, 0, len(selected))
			for _, tmpl := range selected {
				templateNames = append(templateNames, tmpl.Name)
			}

			if err := presets.CreatePreset(name, templateNames); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created preset %s with %d templates\n", name, len(templateNames))
			return nil
		},
	}
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Disable interactive selection")
	return cmd
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
				key := preset.Key
				if strings.TrimSpace(key) == "" {
					key = presets.SluggifyName(preset.Name)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] (%d templates)\n", preset.Name, key, len(preset.Templates))
			}
			return nil
		},
	}
}

func newPresetEditCommand(opts *Options) *cobra.Command {
	var noInteractive bool
	cmd := &cobra.Command{
		Use:   "edit [key] [template1 template2...]",
		Short: "Edit a preset",
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			templateNames := []string{}
			if len(args) > 0 {
				name = args[0]
				if len(args) > 1 {
					templateNames = args[1:]
				}
			}

			items, err := discoverAllTemplates()
			if err != nil {
				return err
			}

			if len(templateNames) > 0 || noInteractive {
				if strings.TrimSpace(name) == "" {
					return fmt.Errorf("preset key or name is required in non-interactive mode")
				}
				index := templates.BuildIndex(items)
				for _, tmpl := range templateNames {
					if _, ok := templates.FindTemplate(index, tmpl); !ok {
						return fmt.Errorf("template not found: %s", tmpl)
					}
				}
				if err := presets.EditPreset(name, templateNames); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Updated preset %s with %d templates\n", name, len(templateNames))
				return nil
			}

			var preset presets.Preset
			if strings.TrimSpace(name) == "" {
				list, err := presets.ListPresets()
				if err != nil {
					return err
				}
				if len(list) == 0 {
					return fmt.Errorf("no presets found")
				}
				preset, err = tui.ShowPresetSelector(list)
				if err != nil {
					if errors.Is(err, tui.ErrCancelled) {
						return nil
					}
					return err
				}
			} else {
				found, ok, err := presets.FindPreset(name)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("preset not found: %s", name)
				}
				preset = found
			}

			selected, err := tui.ShowInteractiveSelector(items, nil, preset.Templates, nil)
			if err != nil {
				if errors.Is(err, tui.ErrCancelled) {
					return nil
				}
				return err
			}
			if len(selected) == 0 {
				return fmt.Errorf("no templates selected")
			}

			templateNames = make([]string, 0, len(selected))
			for _, tmpl := range selected {
				templateNames = append(templateNames, tmpl.Name)
			}

			presetKey := preset.Key
			if strings.TrimSpace(presetKey) == "" {
				presetKey = preset.Name
			}
			if err := presets.EditPreset(presetKey, templateNames); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated preset %s with %d templates\n", preset.Name, len(templateNames))
			return nil
		},
	}
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Disable interactive selection")
	return cmd
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
			if strings.TrimSpace(preset.Key) != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Key: %s\n", preset.Key)
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
		Use:   "delete [key]",
		Short: "Delete a preset",
		RunE: func(cmd *cobra.Command, args []string) error {
			var preset presets.Preset
			if len(args) == 0 {
				list, err := presets.ListPresets()
				if err != nil {
					return err
				}
				if len(list) == 0 {
					return fmt.Errorf("no presets found")
				}
				preset, err = tui.ShowPresetSelector(list)
				if err != nil {
					if errors.Is(err, tui.ErrCancelled) {
						return nil
					}
					return err
				}
			} else {
				found, ok, err := presets.FindPreset(args[0])
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("preset not found: %s", args[0])
				}
				preset = found
			}

			confirm, err := confirmPrompt(cmd, fmt.Sprintf("Delete preset %s?", preset.Name))
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
				return nil
			}

			key := preset.Key
			if strings.TrimSpace(key) == "" {
				key = preset.Name
			}
			if err := presets.DeletePreset(key); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted preset %s\n", preset.Name)
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
		Use:   "use [key]",
		Short: "Generate a .gitignore using a preset",
		RunE: func(cmd *cobra.Command, args []string) error {
			var preset presets.Preset
			interactiveUsed := false
			if len(args) == 0 {
				list, err := presets.ListPresets()
				if err != nil {
					return err
				}
				if len(list) == 0 {
					return fmt.Errorf("no presets found")
				}
				preset, err = tui.ShowPresetSelector(list)
				if err != nil {
					return err
				}
				interactiveUsed = true // Preset selector is interactive
			} else {
				found, ok, err := presets.FindPreset(args[0])
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("preset not found: %s", args[0])
				}
				preset = found
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

			if err := handleExistingOutput(cmd, target, appendMode, force, interactiveUsed, selected); err != nil {
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

func presetKeys() ([]string, error) {
	list, err := presets.ListPresets()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(list))
	for _, preset := range list {
		if strings.TrimSpace(preset.Key) == "" {
			keys = append(keys, presets.SluggifyName(preset.Name))
			continue
		}
		keys = append(keys, preset.Key)
	}
	return keys, nil
}

func presetKeyExists(keys []string, key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, existing := range keys {
		if strings.ToLower(existing) == key {
			return true
		}
	}
	return false
}

func confirmPrompt(cmd *cobra.Command, prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N]: ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	response := strings.ToLower(strings.TrimSpace(line))
	return response == "y" || response == "yes", nil
}
