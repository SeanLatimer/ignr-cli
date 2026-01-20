# ignr

Offline-first gitignore generator CLI tool.

## Overview

`ignr` is a command-line tool that generates `.gitignore` files from templates. It operates offline-first by caching the [github/gitignore](https://github.com/github/gitignore) repository locally, allowing you to generate gitignore files without an internet connection.

## Features

- **Offline-first**: Caches templates locally for fast, offline access
- **Interactive TUI**: Beautiful terminal UI for selecting templates
- **Template Presets**: Save and reuse combinations of templates
- **Auto-suggestions**: Detects your project files and suggests relevant templates
- **Custom Templates**: Add your own custom gitignore templates
- **Template Search**: Fuzzy search through available templates

## Installation

### Homebrew (macOS/Linux)

```bash
brew install seanlatimer/tap/ignr
```

### Scoop (Windows)

```bash
scoop bucket add ignr https://github.com/seanlatimer/scoop-bucket
scoop install ignr
```

### Install Script

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/seanlatimer/ignr-cli/main/scripts/install.sh | sh
```

**Windows PowerShell:**
```powershell
irm https://raw.githubusercontent.com/seanlatimer/ignr-cli/main/scripts/install.ps1 | iex
```

**Windows CMD:**
```cmd
install.bat
```

To install a specific version, set the `VERSION` environment variable:
```bash
VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/seanlatimer/ignr-cli/main/scripts/install.sh | sh
```

### From Source

```bash
git clone https://github.com/seanlatimer/ignr-cli.git
cd ignr-cli
go install ./cmd/ignr
```

### Manual Installation

Download pre-built binaries from the [releases page](https://github.com/seanlatimer/ignr-cli/releases).

## Usage

### Generate a .gitignore

**Interactive mode** (default):
```bash
ignr generate
```

This launches an interactive TUI where you can:
- Search for templates using fuzzy matching
- Select multiple templates
- See suggestions based on your project files (use `--suggest`)

**Non-interactive mode**:
```bash
ignr generate Go Python Node
```

### List Available Templates

```bash
# List all templates
ignr list

# Filter by category
ignr list --category Global
```

### Search Templates

```bash
ignr search python
```

### Update Template Cache

```bash
ignr update
```

### Preset Management

**Create a preset**:
```bash
ignr preset create my-project Go Docker
```

**List presets**:
```bash
ignr preset list
```

**Use a preset**:
```bash
ignr preset use my-project
```

**Interactive preset management**:
```bash
ignr preset
```

This opens an interactive TUI for managing presets.

## Commands

### `ignr generate [template1 template2...]`

Generate a `.gitignore` file from templates.

**Flags:**
- `-o, --output`: Output file path (default: `.gitignore`)
- `--append`: Append to existing file instead of overwriting
- `--no-header`: Skip generator header
- `--force`: Overwrite existing file without prompting
- `--no-interactive`: Disable interactive selection
- `--suggest`: Suggest templates based on repository contents

**Examples:**
```bash
# Interactive selection
ignr generate

# Specific templates
ignr generate Go Python Node

# Custom output location
ignr generate Rust -o .rustignore

# Append to existing file
ignr generate Docker --append
```

### `ignr list`

List available gitignore templates.

**Flags:**
- `--category`: Filter by category (root, Global, community)

### `ignr search <pattern>`

Search templates by name using fuzzy matching.

**Example:**
```bash
ignr search python
```

### `ignr update`

Update the cached gitignore templates from the GitHub repository.

### `ignr preset`

Manage template presets. Run without arguments to open the interactive preset management TUI.

**Subcommands:**
- `create [name] [template1 template2...]`: Create a new preset
- `list`: List all presets
- `show <name>`: Show preset details
- `edit <name>`: Edit a preset
- `delete <name>`: Delete a preset
- `use <name>`: Generate .gitignore from a preset

## Global Flags

- `--config`: Config file path
- `--verbose`: Enable verbose output
- `--quiet`: Suppress non-error output

## Configuration

Configuration is stored in your platform-specific config directory:
- **Windows**: `%APPDATA%\ignr\`
- **Linux/macOS**: `~/.config/ignr/`

### Cache Location

Templates are cached at:
- **Windows**: `%APPDATA%\ignr\cache\github-gitignore`
- **Linux/macOS**: `~/.config/ignr/cache/github-gitignore`

## Custom Templates

You can add your own custom gitignore templates by placing them in:
- **Windows**: `%APPDATA%\ignr\templates\`
- **Linux/macOS**: `~/.config/ignr/templates/`

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
