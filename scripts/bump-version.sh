#!/usr/bin/env bash
# Increment the semantic version in the VERSION file.
#
# Usage: scripts/bump-version.sh [major|minor|patch]
#   patch (default)  0.2.0 -> 0.2.1
#   minor            0.2.0 -> 0.3.0
#   major            0.2.0 -> 1.0.0
set -euo pipefail

cd "$(dirname "$0")/.."

part="${1:-patch}"
current="$(tr -d '[:space:]' < VERSION)"

if [[ ! "$current" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "error: VERSION ('$current') is not semver major.minor.patch" >&2
	exit 1
fi

IFS='.' read -r major minor patch <<< "$current"

case "$part" in
	major) major=$((major + 1)); minor=0; patch=0 ;;
	minor) minor=$((minor + 1)); patch=0 ;;
	patch) patch=$((patch + 1)) ;;
	*) echo "usage: $0 [major|minor|patch]" >&2; exit 1 ;;
esac

new="${major}.${minor}.${patch}"
printf '%s\n' "$new" > VERSION
echo "VERSION: $current -> $new"
