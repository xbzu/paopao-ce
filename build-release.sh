#!/bin/sh
# Build standalone binary archives. This does not build or deploy containers.
# Pass additional Go build tags as the first argument when needed.

TAGS='go_json'

if [ -n "$1" ]; then
  TAGS="$1"
fi

make release CGO_ENABLED=0 TAGS="$TAGS"
