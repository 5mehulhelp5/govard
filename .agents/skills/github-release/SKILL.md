---
name: github-release
description: How to prepare a new release for all projects on Github for govard
risk: unknown
source: local
---

# Prepare Release Workflow

Follow these steps when the user requests to prepare a new release:

## 1. Identify Latest Release and Compare Changes

Identify the latest release tag and compare it with the current `master` branch:

```bash
git fetch origin --tags
LATEST_TAG=$(git describe --tags --abbrev=0)
echo "Latest tag: $LATEST_TAG"

# View summary of changes
git log $LATEST_TAG..origin/master --oneline
git diff $LATEST_TAG..origin/master --stat
```

## 2. Review Detailed Changes

Review the actual code changes since the last release:

```bash
git diff $LATEST_TAG..origin/master
```

## 3. Review Documentation

Review all project documentation (README.md, help files, etc.) to ensure all new features, flags, and changes are accurately documented and up to date.

## 4. Update Version Strings across Projects

Identify and update the version number in all relevant files across all projects.

### Find version occurrences
Replace `X.Y.Z` with the current version and `A.B.C` with the new version:

```bash
# Search for the current version string to identify files to update
grep -r "X.Y.Z" . --exclude-dir=.git
```

### Typical files to update:
- `manifest.json`: `"version": "A.B.C"`
- `package.json`: `"version": "A.B.C"`
- `composer.json`: `"version": "A.B.C"`
- `README.md`: Update version badges and text
- Any HTML version spans (e.g., `<span class="version">vA.B.C</span>`)

## 5. Update CHANGELOG.md

Update the `CHANGELOG.md` file with the new release notes. Include version number, release date, and categorized changes (Added, Changed, Fixed, Removed).

## 6. Prepare GitHub Release Notes

Prepare the GitHub release notes in markdown format.

### Structure
```markdown
## Release vA.B.C

**Date:** Month DD, YYYY

### ✨ New Features
- **Feature:** Description

### 🛠 Improvements
- **Improvement:** Description

### 🐛 Bug Fixes
- **Fix:** Description

**Full Changelog**: https://github.com/USER/REPO/compare/vLATEST...vNEW
```

## 7. Commit and Push Changes

Commit all version-related changes and push to `master` (or a release branch).

```bash
git add .
git commit -m "chore: release vA.B.C"
git push origin master
```

## 8. Create Release Pull Request (If applicable)

If working on a release branch, create a PR to `master`:

```bash
gh pr create --base master --title "chore: release vA.B.C" --body "Release notes content..."
```

## 9. Tag the Release

Create the annotated tag on `master`. **Note:** Pushing a tag is often the trigger for automated release workflows.

```bash
git checkout master
git pull origin master
git tag -a vA.B.C -m "Release vA.B.C"
git push origin vA.B.C
```

## 10. Create GitHub Release Object

Determine if the project has an automated release workflow (e.g., triggered by tag push).

> [!IMPORTANT]
> Check for the existence and configuration of `.github/workflows/release.yml` (or similar) BEFORE running any manual release commands. If an automated workflow is active, DO NOT run `gh release create` as it will cause conflicts or duplicate releases.

### Automated Release (Recommended)
If `.github/workflows/release.yml` exists and is configured with `on: push: tags:`, the release object (including notes and assets) will typically be created automatically by GitHub Actions upon pushing the tag in Step 9.
- **Action:** Monitor progress at `https://github.com/USER/REPO/actions`
- **Verification:** Ensure the Release notes structure in Step 6 is used by the CI pipeline if possible (e.g., by reading `CHANGELOG.md`).

### Manual Release (Fallback Only)
If no automated workflow exists, use the `gh` CLI to create the release object manually.

```bash
gh release create vA.B.C --title "vA.B.C: Release Title" --notes "Release notes..."
```

## When to Use
This skill is applicable for any release preparation or release execution tasks for the govard project on GitHub.
