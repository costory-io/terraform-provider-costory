#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/release.sh <version>

Example:
  scripts/release.sh v0.1.0

This script will:
  1) ensure the working tree is clean
  2) create and verify a signed git tag
  3) push the tag to origin
  4) create a GitHub release with generated notes
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  usage
  exit 1
fi

version="$1"

if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: version must match v<major>.<minor>.<patch> (e.g. v0.1.0)"
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "Error: GitHub CLI (gh) is required."
  exit 1
fi

if ! command -v gpg >/dev/null 2>&1; then
  echo "Error: gpg is required for signed tags."
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: working tree is not clean. Commit or stash changes first."
  exit 1
fi

if git rev-parse "$version" >/dev/null 2>&1; then
  echo "Error: tag '$version' already exists."
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "Error: gh is not authenticated. Run 'gh auth login' first."
  exit 1
fi

echo "Creating signed tag $version..."
git tag -s "$version" -m "Release $version"

echo "Verifying tag signature..."
git tag -v "$version"

echo "Pushing tag to origin..."
git push origin "$version"

echo "Creating GitHub release..."
gh release create "$version" --title "$version" --generate-notes

echo "Release $version completed."
