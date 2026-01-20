package xdginit

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
)

func init() {
	// On Darwin (macOS), prefer ~/.config for CLI tools instead of
	// ~/Library/Application Support, but only if XDG_CONFIG_HOME is not set
	if runtime.GOOS == "darwin" && os.Getenv("XDG_CONFIG_HOME") == "" {
		home := xdg.Home
		xdg.ConfigHome = filepath.Join(home, ".config")
	}
}
