#!/usr/bin/env bash

# description: for updating meta databases, including custom free domains and disposable domains.

set -e
export LC_ALL=C

new=$(mktemp -t emailverifierXXX)

curl --silent https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json | jq -r '.[]' > $new

tmp=$(mktemp -t emailverifierXXX)
cat $new ./disposable.txt \
    | sed '/^$/d' \
    | sed '/./,$!d' \
    | sed -e 's/^ *//' -e 's/ *$//' \
    | awk '{print tolower($0)}' \
    | sort \
    | uniq  > $tmp
mv $tmp ./disposable.txt

sources=$(cat ./sources.txt)
new=$(mktemp -t emailverifierXXX)
for source in $sources; do
    echo "$(curl --silent $source)" >> $new
done;

tmp=$(mktemp -t emailverifierXXX)
cat $new ./free.txt \
    | sed '/^$/d' \
    | sed '/./,$!d' \
    | sed -e 's/^ *//' -e 's/ *$//' \
    | awk '{print tolower($0)}' \
    | sort \
    | uniq \
    | comm -23 - ./disposable.txt > $tmp
mv $tmp ./free.txt

echo 'Complete Updating meta databases!'
