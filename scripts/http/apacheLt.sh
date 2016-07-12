#!/bin/bash
#Output filename: apache_cve-2015-3183
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
printf "CVE-2015-3183 (Version < 2.4.14): `jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion < "0002000400140000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "Total: `jq '.Agents[] | select(.Vendor=="Apache") | .Vendor' $1 | wc -l` \n"
printf '%s\n' '-----------------------'