#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: scripts/bump-version.sh <major|minor|patch>" >&2
}

die() {
  echo "error: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

bump_type="${1:-}"
if [[ $# -ne 1 ]]; then
  usage
  exit 2
fi

case "${bump_type}" in
  major | minor | patch) ;;
  *)
    usage
    exit 2
    ;;
esac

require_command git
require_command perl

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

version_file="internal/buildinfo/version.go"

if [[ -n "$(git status --porcelain)" ]]; then
  die "working tree must be clean"
fi

current_version="$(sed -nE 's/^var Version = "([0-9]+\.[0-9]+\.[0-9]+)"$/\1/p' "${version_file}")"
if [[ -z "${current_version}" ]]; then
  die "could not read semver from ${version_file}"
fi

IFS=. read -r major minor patch <<< "${current_version}"

case "${bump_type}" in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
esac

new_version="${major}.${minor}.${patch}"
tag_name="v${new_version}"
branch_name="release/${tag_name}"

if git rev-parse --verify --quiet "refs/heads/${branch_name}" >/dev/null; then
  die "local branch already exists: ${branch_name}"
fi

if [[ "${DRY_RUN:-}" == "1" ]]; then
  echo "${current_version} -> ${new_version}"
  echo "branch: ${branch_name}"
  echo "commit: chore: bump version to ${new_version}"
  echo "tag: ${tag_name}"
  exit 0
fi

require_command gh
gh auth status >/dev/null 2>&1 || die "gh must be authenticated before creating a PR"

if git ls-remote --exit-code --heads origin "${branch_name}" >/dev/null 2>&1; then
  die "remote branch already exists: ${branch_name}"
fi

if git ls-remote --exit-code --tags origin "${tag_name}" >/dev/null 2>&1; then
  die "remote tag already exists: ${tag_name}"
fi

git switch -c "${branch_name}"
perl -0pi -e "s/var Version = \"\Q${current_version}\E\"/var Version = \"${new_version}\"/" "${version_file}"
git add "${version_file}"
git commit -m "chore: bump version to ${new_version}" -m "Co-authored-by: codex <codex@openai.com>"
git push -u origin "${branch_name}"
gh pr create \
  --base main \
  --head "${branch_name}" \
  --title "chore: bump version to ${new_version}" \
  --body "Bumps export-ua-history to ${new_version}."
