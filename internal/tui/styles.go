package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Package-level styles instance (nil until initialized)
var appStyles *Styles

// Styles holds all application styles using terminal default colors
type Styles struct {
	Primary   color.Color
	Secondary color.Color
	Success   color.Color
	Warning   color.Color
	Error     color.Color
	Subtle    color.Color

	BorderStyle      lipgloss.Style
	SelectedStyle    lipgloss.Style
	SearchInputStyle lipgloss.Style
	FooterStyle      lipgloss.Style
	SubtleStyle      lipgloss.Style
	PresetBadgeStyle lipgloss.Style
	UserBadgeStyle   lipgloss.Style
	SuggestedStyle   lipgloss.Style
	ErrorStyle       lipgloss.Style
	WarningStyle     lipgloss.Style
	SuccessStyle     lipgloss.Style
}

// newStyles creates a new Styles instance using terminal default colors (NoColor)
// All colors use NoColor{} which means "use terminal's default colors"
func newStyles() *Styles {
	// Use NoColor{} everywhere - this tells lipgloss to use the terminal's default colors
	// The terminal itself will provide the colors based on its theme configuration
	noColor := lipgloss.NoColor{}

	return &Styles{
		Primary:   noColor,
		Secondary: noColor,
		Success:   noColor,
		Warning:   noColor,
		Error:     noColor,
		Subtle:    noColor,

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(noColor),

		SelectedStyle: lipgloss.NewStyle().
			Foreground(noColor).
			Bold(true),

		SearchInputStyle: lipgloss.NewStyle().
			Foreground(noColor),

		FooterStyle: lipgloss.NewStyle().
			Foreground(noColor).
			Italic(true),

		SubtleStyle: lipgloss.NewStyle().
			Foreground(noColor),

		PresetBadgeStyle: lipgloss.NewStyle().
			Foreground(noColor).
			Bold(true),

		UserBadgeStyle: lipgloss.NewStyle().
			Foreground(noColor),

		SuggestedStyle: lipgloss.NewStyle().
			Foreground(noColor),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(noColor).
			Bold(true),

		WarningStyle: lipgloss.NewStyle().
			Foreground(noColor).
			Bold(true),

		SuccessStyle: lipgloss.NewStyle().
			Foreground(noColor),
	}
}

// getStyles returns the current styles instance, with fallback for startup
func getStyles() *Styles {
	if appStyles == nil {
		// Initialize styles (compat package will handle background detection)
		return newStyles()
	}
	return appStyles
}
