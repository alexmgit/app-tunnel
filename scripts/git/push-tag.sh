#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 vX.Y.Z"
  echo "Example: $0 v0.1.0"
}

if [[ ${1:-} == "" ]]; then
  usage
  exit 1
fi

tag="$1"
branch=$(git rev-parse --abbrev-ref HEAD)

git tag "$tag"

git push origin "$branch"

git push origin "$tag"
