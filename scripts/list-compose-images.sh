#!/usr/bin/env bash
# Emit unique pinned container images from compose files (excludes :latest).
set -euo pipefail
grep -RhE '^\s+image:[[:space:]]+' \
  --include='docker-compose*.yml' \
  --include='docker-compose*.yaml' \
  . \
  | sed -E 's/.*image:[[:space:]]+//' \
  | sort -u \
  | grep -vE ':latest$'
