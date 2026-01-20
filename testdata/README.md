# Test Data

This directory contains test fixtures and sample data used by the test suite.

## Structure

- `templates/` - Sample gitignore template files for testing template discovery and loading
- `configs/` - Sample configuration files for testing config loading and saving
- `presets/` - Sample preset YAML files for testing preset management

## Usage

Tests should reference these files using relative paths from the test package.
Example:
```go
testdataPath := filepath.Join("testdata", "templates", "Go.gitignore")
```

## Adding New Test Data

When adding new test data:
1. Place files in the appropriate subdirectory
2. Keep file names descriptive
3. Include a comment explaining the purpose if the data is complex
4. Update this README if adding new categories
