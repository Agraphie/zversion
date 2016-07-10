#!/bin/bash
#Output filename: apache_cve-2015-3183
printf $1
printf "\n"
printf "CVE-2015-3183 (Version < 2.4.14): `jq '.Agents[] | select(.Agent=="Apache" and .CanonicalVersion < "0002000400140000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "Total: `jq '.Agents[] | select(.Agent=="Apache") | .Agent' $1 | wc -l`"
printf "\n"
