#!/bin/sh
# Validate the committed tree, never the caller's working copy.
set -eu

fail() { echo "Release check failed: $*" >&2; exit 1; }
[ "$#" -eq 1 ] || fail 'usage: sh tools/check-release.sh vX.Y.Z'
tag=$1
printf '%s\n' "$tag" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' || fail 'expected a stable vX.Y.Z tag'
commit=$(git rev-parse --verify "refs/tags/$tag^{commit}")
version=$(git show "$commit:VERSION")
[ "$tag" = "v$version" ] || fail 'tag does not match VERSION'
[ "$(git log -1 --format=%s "$commit")" = "chore: release $tag" ] || fail "tag must point to chore: release $tag"
[ "$(git diff-tree --no-commit-id --name-only -r "$commit")" = VERSION ] || fail 'release commit must change only VERSION'
echo "Validated $tag at $commit"
