#!/usr/bin/env bash

# description: for updating meta databases, including custom free domains and disposable domains.

set -e
export LC_ALL=C

new=$(mktemp -t emailverifierXXX)

# 1. update disposable domains meta databases
curl --silent https://raw.githubusercontent.com/tompec/disposable-email-domains/main/index.json | jq -r '.[]' > ./disposable.txt

# 2. update free domains meta databases,
curl --silent https://raw.githubusercontent.com/Kikobeats/free-email-domains/refs/heads/master/domains.json | jq -r '.[]' > ./free.txt
sources=$(cat ./free_domain_sources.txt)
new=$(mktemp -t emailverifierXXX)
for source in $sources; do
    echo "$(curl --silent $source)" >> $new
done;

# 3. remove duplicates and sort
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
