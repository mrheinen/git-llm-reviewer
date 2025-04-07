# Release Process for git-llm-reviewer

This document outlines the process for creating and publishing a new release of the git-llm-reviewer tool.

## Prerequisites

Before starting the release process, ensure you have:

- Git configured with proper access to the repository
- Go development environment (version 1.20+)
- Make installed
- Access to GitHub for creating releases (if using GitHub)
- Clean working directory (no uncommitted changes)

## 1. Update Version Information

First, determine the new version number following semantic versioning principles:

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backward compatible manner
- **PATCH** version for backward compatible bug fixes

The version is automatically derived from Git tags and embedded in the binary during build.

## 2. Pre-Release Checklist

Before releasing, complete the following checklist:

- [ ] All tests are passing (`make test`)
- [ ] Documentation is up-to-date (README.md, docs/*)
- [ ] All features for this release are implemented and merged
- [ ] Changelog is updated (see the next section)

## 3. Update Changelog

Create or update the CHANGELOG.md file with details about the new release:

```markdown
# Changelog

## [Unreleased]

## [x.y.z] - YYYY-MM-DD

### Added
- New feature 1
- New feature 2

### Changed
- Change 1
- Change 2

### Fixed
- Bug fix 1
- Bug fix 2

### Removed
- Removed feature 1
- Removed feature 2
```

## 4. Create Git Tag

Tag the release in Git:

```bash
# Ensure you're on the main branch with the latest changes
git checkout main
git pull

# Create an annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z"

# Push the tag to the remote repository
git push origin vX.Y.Z
```

## 5. Build Release Artifacts

Build the release artifacts for all supported platforms:

```bash
# Clean, build, and package all platforms
make release
```

This command will:
1. Clean the build directory
2. Run tests
3. Build binaries for all supported platforms
4. Create ZIP packages with binaries and documentation

The packaged artifacts will be available in the `package/` directory.

## 6. Verify Release Artifacts

Verify that the release artifacts work correctly:

1. Extract each ZIP package
2. Test the binary by running `git-llm-reviewer --version`
3. Verify that the version information is correct
4. Run a basic workflow test on each platform (if possible)

## 7. Publish Release

### On GitHub:

1. Go to the GitHub repository
2. Click on "Releases"
3. Click "Draft a new release"
4. Select the tag version you created
5. Fill in the release title (usually "vX.Y.Z")
6. Copy the relevant section from CHANGELOG.md into the description
7. Attach all ZIP packages from the `package/` directory
8. Click "Publish release"

### On Other Platforms:

If you're distributing through other platforms (e.g., Homebrew, package managers), follow their specific release procedures.

## 8. Update Documentation

After the release is published:

1. Update the installation instructions if necessary
2. Update the download links to point to the new release
3. Announce the release on relevant channels

## 9. Prepare for Next Development Cycle

After the release is complete:

1. Create a new "Unreleased" section in CHANGELOG.md
2. Address any issues that came up during the release process
3. Plan for the next release

## Troubleshooting

### Build Errors

If you encounter build errors during the release process:

1. Verify that you have the correct Go version installed
2. Check for any dependencies that might need updating
3. Run `go mod tidy` to ensure dependencies are correct
4. Try cleaning the build environment with `make clean`

### Version Information Issues

If the version information is not correctly embedded in the binary:

1. Verify that the Git tag is correctly formatted (should be vX.Y.Z)
2. Check that the tag is pushed to the remote repository
3. Ensure the build command includes the correct ldflags

## Release Schedule

We aim to follow these release guidelines:

- **PATCH** releases: As needed for bug fixes
- **MINOR** releases: Every 4-8 weeks with new features
- **MAJOR** releases: When significant changes warrant it, with ample deprecation notices

## Recommended Git Workflow

To maintain a clean Git history and make releases predictable:

1. Develop features on feature branches
2. Use pull requests to merge changes into the main branch
3. Squash commits when merging to maintain a clean history
4. Use semantic commit messages (feat:, fix:, docs:, etc.)
5. Tag releases directly on the main branch

---

This release process helps ensure consistent, high-quality releases of the git-llm-reviewer tool.
