#!/bin/bash
#Output filename: apache_cve-2016-4979
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
printf "CVE-2016-4979 (Version 2.4.18 or 2.4.20): `grep 'Apache","Version":"2.4.18\|Apache","Version":"2.4.20' $1 | jq '.Agents[] | select(.Vendor=="Apache" and (.CanonicalVersion == "0002000400180000" or .CanonicalVersion == "0002000400200000") and .Version != "") | .Version' | wc -l`"
printf "\n"
printf "Total Version >= 2.4.0: `grep 'Apache","Version":"2.4' $1 | jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion >= "0002000400000000" and .CanonicalVersion <= "0002000400230000" and  .Version != "") | .Vendor' | wc -l` \n"
printf "Total: `grep "Apache" $1 | jq '.Agents[] | select(.Vendor=="Apache") | .Vendor' | wc -l` \n"
printf '%s\n' '-----------------------'