# Releasing go-fi

This document explains how to release new versions of go-fi using GitHub Actions.

## Automatic Release Process

### 1. Create a GitHub Release

1. Go to the [Releases page](https://github.com/talinashro/go-fi/releases)
2. Click "Create a new release"
3. Choose a tag (e.g., `v1.0.0`, `v1.1.0`)
4. Write release notes
5. Publish the release

**What happens automatically:**
- The `release.yml` workflow will trigger
- Creates `v1.0.0` tag if this is the first release
- Updates the `latest` tag to point to the new release
- Verifies all tags are created correctly

### 2. Tag-Based Release

Alternatively, you can create a tag directly:

```bash
# Create and push a tag
git tag v1.0.0
git push origin v1.0.0
```

**What happens automatically:**
- The `ci.yml` workflow will trigger
- Runs tests on multiple Go versions
- Builds the project
- Creates `v1.0.0` tag if this is the first release
- Updates the `latest` tag

## Manual Tag Creation

If you need to create tags manually:

1. Go to the [Actions tab](https://github.com/talinashro/go-fi/actions)
2. Select "Manual Tag Creation" workflow
3. Click "Run workflow"
4. Enter the version (e.g., `v1.0.0`)
5. Choose whether to update the `latest` tag
6. Click "Run workflow"

## Tag Strategy

- **`v1.0.0`**: First stable release (created automatically on first release)
- **`latest`**: Always points to the most recent release
- **`v*`**: Semantic versioning tags for each release

## Workflow Files

- **`.github/workflows/ci.yml`**: Main CI/CD pipeline with testing and release
- **`.github/workflows/release.yml`**: Triggered by GitHub releases
- **`.github/workflows/manual-tag.yml`**: Manual tag creation workflow

## Prerequisites

- Repository must have `GITHUB_TOKEN` secret (automatically available)
- Write permissions for the repository (for tag creation)

## Example Release Commands

```bash
# Create a new release
git tag v1.0.0
git push origin v1.0.0

# Or create a GitHub release through the web interface
# Then the workflows will handle tagging automatically
```

The workflows will ensure that:
1. `v1.0.0` is created on the first release
2. `latest` always points to the most recent release
3. All tests pass before tagging
4. Tags are properly verified 