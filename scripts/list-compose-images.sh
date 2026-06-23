#!/usr/bin/env bash
# Emit unique pinned container images from compose files (excludes :latest).
set -euo pipefail
rg 'image:\s+\S+' \
  -g 'docker-compose*.yml' \
  -g 'docker-compose*.yaml' \
  . \
  | sed -E 's/.*image:[[:space:]]+//' \
  | sort -u \
  | grep -vE ':latest$'
