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

# Upstream lists arrive unsorted and with stray whitespace, blank lines,
# comments and mixed case; the HubSpot list is CRLF.
#
# The no-break space substitution is not hypothetical. One entry in the
# tbrianjones list is "atlanticbb.net\u00a0", and under LC_ALL=C
# [[:space:]] does not match U+00A0, so it survived every cleanup this script
# used to do: metadata_free.go shipped that domain with the no-break space
# attached, as a key no lookup could ever match. Turning it into a plain space
# lets the trim below recover the domain, and leaves anything with one in the
# middle for the validation step to reject. The bytes go through $'...' rather
# than a sed escape because BSD sed does not understand \xNN.
#
# Sorting is not cosmetic: the lists are combined with comm below, which needs
# sorted inputs and silently misreports without them.
normalise() {
    sed $'s/\302\240/ /g' \
        | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' \
        | sed -e '/^$/d' -e '/^#/d' \
        | awk '{print tolower($0)}' \
        | sort -u
}

# Whatever is left that is not shaped like a hostname says something went
# wrong upstream or in transit -- a truncated line, an error page, two sources
# concatenated without a newline between them. Drop those, but say so: a
# silent drop is how a source turning into an error page looks like a source
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
# the data. The floors catch an empty or truncated response, which would
# otherwise be published as "nothing is disposable" and let every disposable
# domain through as a free one.
#
# The free ceiling catches the opposite failure, which is the one that actually
# happened: a source silently starts including disposable domains and the list
# doubles. free.txt is a slowly growing set of real mailbox providers, so a
# jump past this is a change in what a source publishes, not organic growth.
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
#    This URL is also in constants.go as disposableDataURL, where
#    EnableAutoUpdateDisposable fetches it at runtime. Keep the two identical.
#    They were not: this script built the list from tompec while the library
#    refreshed it from here, and the two share only 37585 entries out of
#    tompec's 133602. Since updateDisposableDomains deletes whatever the
#    fetched list omits, enabling auto-update dropped 96017 baked-in domains
#    and added 37679, leaving 75264.
fetch https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains.json \
    | jq -r '.[]' | normalise | keep_valid_domains > "$workdir/disposable.txt"
require_line_count "$workdir/disposable.txt" 50000 - "disposable domains"

# 2. update free domains meta databases, from every source in
#    free_domain_sources.txt plus the domains vendored in free_domains_extra.txt
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
cat ./free_domains_extra.txt >> "$workdir/free_raw.txt"

# 3. normalise, then drop anything already known to be disposable. The free
#    sources are signup blocklists rather than provider directories -- HubSpot
#    publishes theirs under Marketing/Lead-Capture -- so they carry throwaway
#    domains alongside the real providers and this subtraction is what
#    separates the two.
normalise < "$workdir/free_raw.txt" \
    | keep_valid_domains \
    | comm -23 - "$workdir/disposable.txt" > "$workdir/free.txt"
require_line_count "$workdir/free.txt" 3000 6000 "free domains"

# 4. publish, only now that both lists are known good. Writing through the
#    existing files keeps their permissions; mv would carry mktemp's 0600 over,
#    leaving free.txt unreadable to other users in the working tree of whoever
#    ran this. Never in the repository -- git records only the exec bit -- but
#    it does happen locally.
cat "$workdir/disposable.txt" > ./disposable.txt
cat "$workdir/free.txt" > ./free.txt

echo 'Complete Updating meta databases!'
