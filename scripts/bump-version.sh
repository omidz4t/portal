#!/usr/bin/env bash
# Semver bump from conventional commits. No extra packages.
#
#   fix: / perf: / revert:     → patch
#   feat:                      → minor
#   type!: or BREAKING CHANGE  → major
#   chore/docs/ci/test/style   → no bump (unless --include-chore)
#
# Usage:
#   scripts/bump-version.sh
#   scripts/bump-version.sh --dry-run
#   scripts/bump-version.sh --bump patch|minor|major
#   scripts/bump-version.sh --force 1.2.3
#   scripts/bump-version.sh --commit --tag --changelog
#   scripts/bump-version.sh --from 0.1.0 --include-chore
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION_FILE="${VERSION_FILE:-VERSION}"
CHANGELOG_FILE="${CHANGELOG_FILE:-CHANGELOG.md}"
DRY_RUN=0
DO_TAG=0
DO_COMMIT=0
DO_CHANGELOG=0
INCLUDE_CHORE=0
SKIP_CI=0
FORCE_VERSION=""
FORCE_BUMP=""
FROM_VERSION=""

IN_CI=0
if [[ "${CI:-}" == "true" ]]; then
	IN_CI=1
	SKIP_CI=1
fi

usage() {
	sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
	exit "${1:-0}"
}

while [[ $# -gt 0 ]]; do
	case "$1" in
		-h|--help) usage 0 ;;
		-n|--dry-run) DRY_RUN=1; shift ;;
		--tag) DO_TAG=1; shift ;;
		--no-tag) DO_TAG=0; shift ;;
		--commit) DO_COMMIT=1; shift ;;
		--no-commit) DO_COMMIT=0; shift ;;
		--changelog) DO_CHANGELOG=1; shift ;;
		--release) DO_CHANGELOG=1; DO_TAG=1; DO_COMMIT=1; shift ;;
		--include-chore) INCLUDE_CHORE=1; shift ;;
		--skip-ci) SKIP_CI=1; shift ;;
		--from)
			FROM_VERSION="${2:-}"
			[[ -n "$FROM_VERSION" ]] || { echo "error: --from needs X.Y.Z" >&2; exit 1; }
			shift 2
			;;
		--bump)
			FORCE_BUMP="${2:-}"
			case "$FORCE_BUMP" in
				patch|minor|major) ;;
				*) echo "error: --bump needs patch|minor|major" >&2; exit 1 ;;
			esac
			shift 2
			;;
		--force)
			FORCE_VERSION="${2:-}"
			[[ -n "$FORCE_VERSION" ]] || { echo "error: --force needs X.Y.Z" >&2; exit 1; }
			shift 2
			;;
		*)
			echo "error: unknown option: $1" >&2
			usage 1
			;;
	esac
done

is_semver() {
	[[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

strip_v() {
	local v="$1"
	v="${v#v}"
	v="${v#V}"
	printf '%s' "$v"
}

latest_tag() {
	git tag --list 'v*' --sort=-v:refname 2>/dev/null | while read -r tag; do
		local ver
		ver="$(strip_v "$tag")"
		if is_semver "$ver"; then
			printf '%s' "$tag"
			return
		fi
	done
}

current_version() {
	if [[ -n "$FROM_VERSION" ]]; then
		printf '%s' "$(strip_v "$FROM_VERSION")"
		return
	fi
	if [[ -f "$VERSION_FILE" ]]; then
		local f
		f="$(tr -d '[:space:]' <"$VERSION_FILE")"
		f="$(strip_v "$f")"
		if is_semver "$f"; then
			printf '%s' "$f"
			return
		fi
	fi
	local tag
	tag="$(latest_tag)"
	if [[ -n "$tag" ]]; then
		printf '%s' "$(strip_v "$tag")"
		return
	fi
	printf '0.0.0'
}

commit_range() {
	if [[ -n "$FROM_VERSION" ]]; then
		printf 'HEAD'
		return
	fi
	local tag
	tag="$(latest_tag)"
	if [[ -n "$tag" ]]; then
		printf '%s..HEAD' "$tag"
	else
		printf 'HEAD'
	fi
}

# Prints: major|minor|patch|none
detect_bump() {
	local range msgs bodies
	range="$(commit_range)"

	if [[ "$range" == "HEAD" ]]; then
		msgs="$(git log --pretty=%s --no-merges 2>/dev/null || true)"
		bodies="$(git log --pretty=%b --no-merges 2>/dev/null || true)"
	else
		msgs="$(git log --pretty=%s --no-merges "$range" 2>/dev/null || true)"
		bodies="$(git log --pretty=%b --no-merges "$range" 2>/dev/null || true)"
	fi

	if [[ -z "$msgs" ]]; then
		printf 'none'
		return
	fi

	if printf '%s\n' "$msgs" | grep -qE '^[a-zA-Z]+(\([^)]+\))?!:'; then
		printf 'major'
		return
	fi
	if printf '%s\n' "$bodies" | grep -qiE '^BREAKING[ -]CHANGE:'; then
		printf 'major'
		return
	fi
	if printf '%s\n' "$msgs" | grep -qE '^feat(\([^)]+\))?:'; then
		printf 'minor'
		return
	fi
	if printf '%s\n' "$msgs" | grep -qE '^(fix|perf|revert)(\([^)]+\))?:'; then
		printf 'patch'
		return
	fi
	if [[ "$INCLUDE_CHORE" -eq 1 ]] &&
		printf '%s\n' "$msgs" | grep -qE '^(chore|docs|style|refactor|test|build|ci)(\([^)]+\))?:'; then
		printf 'patch'
		return
	fi
	printf 'none'
}

bump_version() {
	local cur="$1" kind="$2"
	local major minor patch
	IFS=. read -r major minor patch <<<"$cur"

	case "$kind" in
		major) major=$((major + 1)); minor=0; patch=0 ;;
		minor) minor=$((minor + 1)); patch=0 ;;
		patch) patch=$((patch + 1)) ;;
		none) printf '%s' "$cur"; return ;;
		*) echo "error: unknown bump: $kind" >&2; exit 1 ;;
	esac
	printf '%s.%s.%s' "$major" "$minor" "$patch"
}

write_changelog() {
	local version="$1" kind="$2" range="$3"
	local date subjects
	date="$(date -u +%Y-%m-%d)"
	if [[ "$range" == "HEAD" ]]; then
		subjects="$(git log --pretty=%s --no-merges 2>/dev/null || true)"
	else
		subjects="$(git log --pretty=%s --no-merges "$range" 2>/dev/null || true)"
	fi

	local breaking feats fixes other
	breaking=""; feats=""; fixes=""; other=""
	while IFS= read -r line; do
		[[ -z "$line" ]] && continue
		if [[ "$line" =~ ^[a-zA-Z]+(\([^\)]+\))?!: ]]; then
			breaking+="- ${line}"$'\n'
		elif [[ "$line" =~ ^feat(\([^\)]+\))?: ]]; then
			feats+="- ${line}"$'\n'
		elif [[ "$line" =~ ^(fix|perf|revert)(\([^\)]+\))?: ]]; then
			fixes+="- ${line}"$'\n'
		elif [[ "$INCLUDE_CHORE" -eq 1 ]]; then
			other+="- ${line}"$'\n'
		fi
	done <<<"$subjects"

	{
		echo "## ${version} (${date})"
		echo
		echo "Release type: **${kind}**"
		echo
		if [[ -n "$breaking" ]]; then
			echo "### Breaking changes"
			echo
			printf '%s' "$breaking"
			echo
		fi
		if [[ -n "$feats" ]]; then
			echo "### Features"
			echo
			printf '%s' "$feats"
			echo
		fi
		if [[ -n "$fixes" ]]; then
			echo "### Fixes"
			echo
			printf '%s' "$fixes"
			echo
		fi
		if [[ -n "$other" ]]; then
			echo "### Other"
			echo
			printf '%s' "$other"
			echo
		fi
	} >"${CHANGELOG_FILE}.section"

	if [[ -f "$CHANGELOG_FILE" ]]; then
		{
			echo "# Changelog"
			echo
			cat "${CHANGELOG_FILE}.section"
			# drop an existing leading title so we do not stack two
			sed '1{/^# Changelog/d;}' "$CHANGELOG_FILE" | sed '1{/^$/d;}'
		} >"${CHANGELOG_FILE}.new"
		mv "${CHANGELOG_FILE}.new" "$CHANGELOG_FILE"
	else
		{
			echo "# Changelog"
			echo
			cat "${CHANGELOG_FILE}.section"
		} >"$CHANGELOG_FILE"
	fi
	rm -f "${CHANGELOG_FILE}.section"
}

if [[ -n "$FORCE_VERSION" && -n "$FORCE_BUMP" ]]; then
	echo "error: use only one of --force or --bump" >&2
	exit 1
fi

if [[ -n "$FORCE_VERSION" ]]; then
	FORCE_VERSION="$(strip_v "$FORCE_VERSION")"
	if ! is_semver "$FORCE_VERSION"; then
		echo "error: --force must be X.Y.Z (got $FORCE_VERSION)" >&2
		exit 1
	fi
	CUR="$(current_version)"
	NEW="$FORCE_VERSION"
	KIND="force"
elif [[ -n "$FORCE_BUMP" ]]; then
	CUR="$(current_version)"
	KIND="$FORCE_BUMP"
	NEW="$(bump_version "$CUR" "$KIND")"
else
	CUR="$(current_version)"
	KIND="$(detect_bump)"
	NEW="$(bump_version "$CUR" "$KIND")"
fi

RANGE="$(commit_range)"
echo "range:   $RANGE"
echo "current: $CUR"
echo "bump:    $KIND"
echo "next:    $NEW"

if [[ "$KIND" == "none" && -z "$FORCE_VERSION" ]]; then
	echo "no releasable conventional commits since last tag — version unchanged"
	if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
		{
			echo "bumped=false"
			echo "version=$CUR"
			echo "tag=v$CUR"
		} >>"$GITHUB_OUTPUT"
	fi
	exit 0
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
	echo "(dry-run) would write $VERSION_FILE → $NEW"
	if [[ "$DO_CHANGELOG" -eq 1 ]]; then
		echo "(dry-run) would prepend $CHANGELOG_FILE"
	fi
	if [[ "$DO_COMMIT" -eq 1 ]]; then
		echo "(dry-run) would commit chore(release): $NEW"
	fi
	if [[ "$DO_TAG" -eq 1 ]]; then
		echo "(dry-run) would tag v$NEW"
	fi
	if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
		{
			echo "bumped=true"
			echo "version=$NEW"
			echo "tag=v$NEW"
		} >>"$GITHUB_OUTPUT"
	fi
	exit 0
fi

printf '%s\n' "$NEW" >"$VERSION_FILE"
echo "wrote $VERSION_FILE"

if [[ "$DO_CHANGELOG" -eq 1 ]]; then
	write_changelog "$NEW" "$KIND" "$RANGE"
	echo "wrote $CHANGELOG_FILE"
fi

if [[ "$DO_COMMIT" -eq 1 ]]; then
	git add -- "$VERSION_FILE"
	[[ "$DO_CHANGELOG" -eq 1 && -f "$CHANGELOG_FILE" ]] && git add -- "$CHANGELOG_FILE"
	if git diff --cached --quiet; then
		echo "nothing to commit"
	else
		MSG="chore(release): $NEW"
		[[ "$SKIP_CI" -eq 1 ]] && MSG="$MSG [skip ci]"
		if [[ "$IN_CI" -eq 1 ]]; then
			git commit --no-gpg-sign -m "$MSG"
			echo "committed $MSG"
		elif git commit -S -m "$MSG" 2>/dev/null; then
			echo "committed $MSG (signed)"
		else
			git commit -m "$MSG"
			echo "committed $MSG"
		fi
	fi
fi

if [[ "$DO_TAG" -eq 1 ]]; then
	TAG="v$NEW"
	if git rev-parse "$TAG" >/dev/null 2>&1; then
		echo "error: tag $TAG already exists" >&2
		exit 1
	fi
	if [[ "$IN_CI" -eq 1 ]]; then
		git tag -a "$TAG" -m "TGPORTAL $TAG"
		echo "created tag $TAG"
	elif git tag -s "$TAG" -m "TGPORTAL $TAG" 2>/dev/null; then
		echo "created signed tag $TAG"
	else
		git tag -a "$TAG" -m "TGPORTAL $TAG"
		echo "created annotated tag $TAG"
	fi
	echo "push with: git push origin HEAD $TAG"
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
	{
		echo "bumped=true"
		echo "version=$NEW"
		echo "tag=v$NEW"
	} >>"$GITHUB_OUTPUT"
fi
