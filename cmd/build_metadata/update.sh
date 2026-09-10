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

workdir=$(mktemp -d -t emailverifierXXXXXX)
trap 'rm -rf "$workdir"' EXIT

# --fail matters: without it curl exits 0 on an HTTP error and writes the error
# body to stdout, which then gets published as domain data.
fetch() {
    curl --silent --show-error --fail --location "$1"
}

# Upstream lists carry stray whitespace, blank lines, comments and mixed case,
# and HubSpot's is CRLF. One tbrianjones entry ends in U+00A0, which LC_ALL=C
# [[:space:]] does not match, so it needs a substitution of its own -- as
# literal bytes through $'...', since \xNN is not portable.
#
# The sort is required, not tidiness: comm below needs sorted inputs, and
# without them GNU comm aborts the run while BSD comm publishes a wrong list.
normalise() {
    sed $'s/\302\240/ /g' \
        | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' \
        | sed -e '/^$/d' -e '/^#/d' \
        | awk '{print tolower($0)}' \
        | sort -u
}

# Anything not shaped like a hostname means the fetch or the source went wrong.
# Report what gets dropped: shrinking in silence looks the same as a source
# that simply got smaller.
keep_valid_domains() {
    awk '
        /^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/ { print; next }
        {
            invalid++
            if (invalid <= 10) printf "  discarding malformed entry: %s\n", $0 > "/dev/stderr"
        }
        END { if (invalid) printf "discarded %d malformed entries\n", invalid > "/dev/stderr" }
    '
}

# A count outside these bounds means the fetch returned something other than
# the data. The floors catch an empty response, which would otherwise publish
# "nothing is disposable" and let every disposable domain through as free. The
# free ceiling catches the failure that did happen: a source starting to merge
# disposable blocklists into itself, which doubled the list.
require_line_count() {
    local file=$1 min=$2 max=$3 label=$4
    local count
    count=$(wc -l < "$file")
    if [ "$count" -lt "$min" ]; then
        echo "$label: got $count entries, expected at least $min -- refusing to publish" >&2
        exit 1
    fi
    if [ "$max" != "-" ] && [ "$count" -gt "$max" ]; then
        echo "$label: got $count entries, expected at most $max -- refusing to publish" >&2
        exit 1
    fi
    echo "$label: $count entries"
}

# 1. update disposable domains meta databases.
#
#    Same URL as disposableDataURL in constants.go, where
#    EnableAutoUpdateDisposable fetches it at runtime. They have to match:
#    updateDisposableDomains deletes whatever the fetched list omits, so a
#    different source here means auto-update replaces most of the baked-in
#    list. TestUpdateScriptFetchesTheDisposableDataURL holds them together.
fetch https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains.json \
    | jq -r '.[]' | normalise | keep_valid_domains > "$workdir/upstream_disposable.txt"
require_line_count "$workdir/upstream_disposable.txt" 50000 - "disposable domains"

#    Then remove disposable_allowlist.txt, domains the upstream list calls
#    disposable and which are not. Here rather than only at runtime because the
#    generated map is what a caller gets without EnableAutoUpdateDisposable,
#    and because step 3 subtracts this file, so a domain taken out here stays
#    in free.txt rather than being reported as neither. An entry upstream has
#    since dropped is dead weight, hence the check.
normalise < ./disposable_allowlist.txt > "$workdir/allowlist.txt"
stale=$(comm -23 "$workdir/allowlist.txt" "$workdir/upstream_disposable.txt")
if [ -n "$stale" ]; then
    echo "disposable_allowlist.txt lists domains the upstream list no longer has:" >&2
    echo "$stale" | sed 's/^/  /' >&2
    echo "remove them from disposable_allowlist.txt -- refusing to publish" >&2
    exit 1
fi
comm -23 "$workdir/upstream_disposable.txt" "$workdir/allowlist.txt" > "$workdir/disposable.txt"
echo "disposable allowlist: $(wc -l < "$workdir/allowlist.txt") entries removed"

# 2. update free domains meta databases, from every source in
#    free_domain_sources.txt plus the domains vendored in free_domains_extra.txt
while read -r source; do
    [ -n "$source" ] || continue
    fetch "$source" >> "$workdir/free_raw.txt"
    # Not every source ends with a newline -- the tbrianjones list does not --
    # so terminate each one, or its last domain merges with the first line of
    # the next. normalise drops the blank lines this leaves.
    echo >> "$workdir/free_raw.txt"
done < ./free_domain_sources.txt
cat ./free_domains_extra.txt >> "$workdir/free_raw.txt"

# 3. normalise, then drop anything known to be disposable. The free sources are
#    signup blocklists rather than provider directories -- HubSpot publishes
#    theirs under Marketing/Lead-Capture -- so they carry throwaway domains
#    alongside the real ones, and this subtraction is what separates them.
normalise < "$workdir/free_raw.txt" \
    | keep_valid_domains \
    | comm -23 - "$workdir/disposable.txt" > "$workdir/free.txt"
require_line_count "$workdir/free.txt" 3000 6000 "free domains"

# 4. publish, only now that both lists are known good. cat rather than mv so
#    the targets keep their permissions: mv carries mktemp's 0600 across,
#    leaving free.txt unreadable to others in the working tree of whoever ran
#    it. Not in the repository -- git records only the exec bit.
cat "$workdir/disposable.txt" > ./disposable.txt
cat "$workdir/free.txt" > ./free.txt

echo 'Complete Updating meta databases!'
