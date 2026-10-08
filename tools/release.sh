#!/bin/sh
# Prepare a local release. Publishing is a separate, explicit git push.
set -eu

fail() { echo "Release failed: $*" >&2; exit 1; }
cd "$(dirname "$0")/.."
[ "$#" -eq 1 ] || fail 'usage: sh tools/release.sh X.Y.Z'
version=$1
printf '%s\n' "$version" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' || fail 'expected a stable X.Y.Z version'
[ "$(git branch --show-current)" = main ] || fail 'release from main'
[ -z "$(git status --porcelain)" ] || fail 'commit all pending changes first'
previous=$(cat VERSION)
[ "$version" != "$previous" ] || fail 'version is unchanged'
[ "$(printf '%s\n%s\n' "$previous" "$version" | sort -V | tail -1)" = "$version" ] || fail 'version must increase'
tag="v$version"
if git show-ref --verify --quiet "refs/tags/$tag"; then
  fail "$tag already exists"
fi

printf '%s\n' "$version" > VERSION
git add -- VERSION
git commit -m "chore: release $tag"
git tag -a "$tag" -m "Release $tag"
sh tools/check-release.sh "$tag"
printf '\nReady to publish: git push --atomic origin main %s\n' "$tag"
