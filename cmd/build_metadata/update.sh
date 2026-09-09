#!/usr/bin/env bash

# description: for updating meta databases, including custom free domains and disposable domains.
#
# Refreshes the .txt lists only. The generated metadata_*.go files are built by
# `go run main.go` in this directory, which runs this script first -- editing a
# .txt list on its own leaves the compiled maps stale.
#
# Can be run from anywhere; paths resolve relative to the script.

set -euo pipefail
export LC_ALL=C

cd "$(dirname "${BASH_SOURCE[0]}")"

# A list this small means the fetch returned something other than the data. Both
# sources are an order of magnitude larger; these floors exist to catch an empty
# or truncated response, which would otherwise be published as "nothing is
# disposable" and let every disposable domain through as a free one.
readonly MIN_DISPOSABLE=50000
readonly MIN_FREE=3000

workdir=$(mktemp -d -t emailverifierXXXXXX)
trap 'rm -rf "$workdir"' EXIT

# --fail matters: without it curl exits 0 on an HTTP error and writes the error
# body to stdout, which then gets published as domain data.
fetch() {
    curl --silent --show-error --fail --location "$1"
}

require_min_lines() {
    local file=$1 min=$2 label=$3
    local count
    count=$(wc -l < "$file")
    if [ "$count" -lt "$min" ]; then
        echo "$label: got $count entries, expected at least $min -- refusing to publish" >&2
        exit 1
    fi
    echo "$label: $count entries"
}

# 1. update disposable domains meta databases
#    Sorted here rather than trusting the upstream order: step 3 hands this file
#    to comm, which needs both inputs sorted and misreports if they are not.
fetch https://raw.githubusercontent.com/tompec/disposable-email-domains/main/index.json \
    | jq -r '.[]' \
    | sort -u > "$workdir/disposable.txt"
require_min_lines "$workdir/disposable.txt" "$MIN_DISPOSABLE" "disposable domains"

# 2. update free domains meta databases, from the primary source plus anything
#    listed in free_domain_sources.txt
fetch https://raw.githubusercontent.com/Kikobeats/free-email-domains/refs/heads/master/domains.json \
    | jq -r '.[]' > "$workdir/free_raw.txt"
while read -r source; do
    [ -n "$source" ] || continue
    fetch "$source" >> "$workdir/free_raw.txt"
    # Not every source ends with a newline -- the tbrianjones list does not --
    # so terminate each one explicitly, or its last domain merges with the
    # first line of whatever is appended next. This is what the original
    # script's echo "$(curl ...)" was doing. normalise drops the blank lines
    # it leaves behind.
    echo >> "$workdir/free_raw.txt"
done < ./free_domain_sources.txt

# 3. normalise, remove duplicates and sort, then drop anything already known to
#    be disposable. Whitespace is trimmed before blank lines are dropped so that
#    a whitespace-only line does not survive as an empty entry.
sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' "$workdir/free_raw.txt" \
    | sed '/^$/d' \
    | awk '{print tolower($0)}' \
    | sort -u \
    | comm -23 - "$workdir/disposable.txt" > "$workdir/free.txt"
require_min_lines "$workdir/free.txt" "$MIN_FREE" "free domains"

# 4. publish, only now that both lists are known good. Writing through the
#    existing files keeps their permissions; mv would carry mktemp's 0600 over,
#    leaving free.txt unreadable to other users in the working tree of whoever
#    ran this. Never in the repository -- git records only the exec bit -- but
#    it does happen locally.
cat "$workdir/disposable.txt" > ./disposable.txt
cat "$workdir/free.txt" > ./free.txt

echo 'Complete Updating meta databases!'
