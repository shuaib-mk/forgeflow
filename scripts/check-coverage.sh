#!/usr/bin/env sh
set -eu

profile="${1:-coverage.out}"
minimum="${2:-70}"

if [ ! -f "$profile" ]; then
  printf 'Coverage profile not found: %s\n' "$profile" >&2
  exit 1
fi

total="$(go tool cover -func="$profile" | awk '/^total:/ { gsub("%", "", $3); print $3 }')"
if [ -z "$total" ]; then
  printf 'Could not read total coverage from %s\n' "$profile" >&2
  exit 1
fi

printf 'Coverage: %s%% (required: %s%%)\n' "$total" "$minimum"
awk -v actual="$total" -v required="$minimum" 'BEGIN { exit !(actual + 0 >= required + 0) }'
