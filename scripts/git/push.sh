#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 \"commit message\""
  echo "Example: $0 \"feat: add auth token\""
}

if [[ ${1:-} == "" ]]; then
  usage
  exit 1
fi

msg="$1"
branch=$(git rev-parse --abbrev-ref HEAD)

if [[ -n "$(git status --porcelain)" ]]; then
  git add -A
  git commit -m "$msg"
fi

git push origin "$branch"
