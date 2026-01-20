package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.seanlatimer.dev/ignr/internal/cache"
	"go.seanlatimer.dev/ignr/internal/config"
	"go.seanlatimer.dev/ignr/internal/presets"
	"go.seanlatimer.dev/ignr/internal/templates"
	"go.seanlatimer.dev/ignr/internal/tui"
)

func newGenerateCommand(opts *Options) *cobra.Command {
	var output string
	var appendMode bool
	var noHeader bool
	var force bool
	var noInteractive bool
	var suggest bool

	cmd := &cobra.Command{
		Use:   "generate [template1 template2...]",
		Short: "Generate a .gitignore from templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cachePath, err := cache.InitializeCache()
			if err != nil {
				return err
			}

			items, err := templates.DiscoverTemplates(cachePath)
			if err != nil {
				return err
			}
			userPath, err := config.GetUserTemplatePath()
			if err != nil {
				return err
			}
			userItems, err := templates.DiscoverUserTemplates(userPath)
			if err != nil {
				return err
			}
			items = append(items, userItems...)

			presetList, err := presets.ListPresets()
			if err != nil {
				return err
			}

			suggested := []string{}
			if suggest && len(args) == 0 && !noInteractive {
				detected, err := presets.DetectFiles(".")
				if err != nil {
					return err
				}
				suggested, err = presets.SuggestTemplates(detected)
				if err != nil {
					return err
				}
			}

			selected, interactiveUsed, err := selectTemplates(args, items, presetList, suggested, noInteractive)
			if err != nil {
				if errors.Is(err, tui.ErrCancelled) {
					return nil
				}
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Generated %s with %d templates\n", target, len(selected))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: .gitignore)")
	cmd.Flags().BoolVar(&appendMode, "append", false, "Append to existing file instead of overwrite")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Skip generator header")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file without prompt")
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Disable interactive selection")
	cmd.Flags().BoolVar(&suggest, "suggest", false, "Suggest templates based on repo contents")
	return cmd
}

func selectTemplates(args []string, items []templates.Template, presetList []presets.Preset, suggested []string, noInteractive bool) ([]templates.Template, bool, error) {
	if len(args) > 0 || noInteractive {
		index := templates.BuildIndex(items)
		selected := make([]templates.Template, 0, len(args))
		for _, name := range args {
			t, ok := templates.FindTemplate(index, name)
			if !ok {
				return nil, false, fmt.Errorf("template not found: %s", name)
			}
			selected = append(selected, t)
		}
		return selected, false, nil
	}

	selected, err := tui.ShowInteractiveSelector(items, presetList, nil, suggested)
	return selected, true, err
}

func resolveOutputPath(output string) (string, error) {
	if strings.TrimSpace(output) != "" {
		return output, nil
	}

	cfg, err := config.LoadConfig()
	if err == nil && strings.TrimSpace(cfg.DefaultOutput) != "" {
		return cfg.DefaultOutput, nil
	}

	return filepath.Join(".", ".gitignore"), nil
}

func handleExistingOutput(cmd *cobra.Command, path string, appendMode, force, interactive bool, templates []templates.Template) error {
	if appendMode || force {
		return nil
	}
	if !fileExists(path) {
		return nil
	}

	if !interactive {
		return fmt.Errorf("output file exists: %s (use --force or --append)", path)
	}

	confirm, err := tui.ConfirmOverwrite(path, templates)
	if err != nil {
		if errors.Is(err, tui.ErrCancelled) {
			return tui.ErrCancelled
		}
		return err
	}
	if !confirm {
		return tui.ErrCancelled
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeOutput(path, content string, appendMode, force bool) error {
	if appendMode {
		return appendToFile(path, content)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func appendToFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := file.WriteString(content)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}
