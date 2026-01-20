# Release Setup and Process

This document describes how to set up and perform releases for ignr.

## Initial Setup (One-Time)

Before making your first release, you need to create the Homebrew tap and Scoop bucket repositories.

### 1. Create Homebrew Tap Repository

1. Create a new repository on GitHub: `seanlatimer/homebrew-tap`
2. Initialize it with an empty `Formula/` directory:
   ```bash
   mkdir -p Formula
   touch Formula/.gitkeep
   git init
   git add Formula/.gitkeep
   git commit -m "Initialize homebrew tap"
   git remote add origin https://github.com/seanlatimer/homebrew-tap.git
   git push -u origin main
   ```

### 2. Create Scoop Bucket Repository

1. Create a new repository on GitHub: `seanlatimer/scoop-bucket`
2. Initialize it with an empty `bucket/` directory:
   ```bash
   mkdir -p bucket
   touch bucket/.gitkeep
   git init
   git add bucket/.gitkeep
   git commit -m "Initialize scoop bucket"
   git remote add origin https://github.com/seanlatimer/scoop-bucket.git
   git push -u origin main
   ```

## Release Process

### Creating a Release

1. **Ensure all changes are committed and pushed:**
   ```bash
   git add .
   git commit -m "Your changes"
   git push
   ```

2. **Create and push a git tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions will automatically:**
   - Run the full test suite
   - Build binaries for Windows, Linux, and macOS (amd64 and arm64)
   - Create a GitHub release with all binaries
   - Update the Homebrew tap repository with a new formula
   - Update the Scoop bucket repository with a new manifest

### Verifying the Release

1. Check the [GitHub releases page](https://github.com/seanlatimer/ignr-cli/releases)
2. Verify binaries are available for all platforms
3. Test Homebrew installation:
   ```bash
   brew install seanlatimer/tap/ignr
   ignr --version
   ```
4. Test Scoop installation:
   ```bash
   scoop bucket add ignr https://github.com/seanlatimer/scoop-bucket
   scoop install ignr
   ignr --version
   ```

## Version Numbering

- Use [semantic versioning](https://semver.org/): `MAJOR.MINOR.PATCH`
- Tag format: `v1.0.0` (must start with `v`)
- Examples: `v1.0.0`, `v1.2.3`, `v2.0.0`

## Testing Locally

Before creating a release tag, you can test the goreleaser configuration locally:

```bash
# Dry-run (doesn't publish anything)
goreleaser release --snapshot --skip-publish

# Build for specific platform
goreleaser build --snapshot --single-target
```

## Troubleshooting

### Release workflow fails

- Check that tests pass locally
- Verify `.goreleaser.yaml` is valid: `goreleaser check`
- Ensure git tags are pushed: `git push --tags`

### Homebrew/Scoop update fails

- Verify repositories exist and are public
- Check that repositories have write access
- Review GitHub Actions logs for specific errors
