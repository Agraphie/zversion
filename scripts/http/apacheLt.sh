#!/bin/bash
#Output filename: apache_cve-2015-3183
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
printf "CVE-2015-3183 (2.4.0 <= Version < 2.4.14): `grep "Apache" $1 | jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion < "0002000400140000" and .CanonicalVersion >= "0002000400000000" and .Version != "") | .Version' | wc -l`"
printf "\n"
printf "Total Version >= 2.4.0: `grep "Apache" $1 | jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion >= "0002000400000000" and .Version != "") | .Vendor' | wc -l` \n"
printf "Total: `grep "Apache" $1 | jq '.Agents[] | select(.Vendor=="Apache") | .Vendor' | wc -l` \n"
printf '%s\n' '-----------------------'