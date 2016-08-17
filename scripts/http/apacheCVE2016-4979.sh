#!/bin/bash
#Output filename: apache_cve-2016-4979
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "Apache" $1 > $tmpName

vulnerable=`grep 'Apache","Version":"2.4.18\|Apache","Version":"2.4.20' $tmpName | jq '.Agents[] | select(.Vendor=="Apache" and (.CanonicalVersion == "0002000400180000" or .CanonicalVersion == "0002000400200000") and .Version != "") | .Version' | wc -l`
total24=$(grep 'Apache","Version":"2.4' $tmpName | jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion >= "0002000400000000" and .CanonicalVersion <= "0002000400230000" and  .Version != "") | .Vendor' | wc -l)

printf "%s\n" "CVE-2016-4979 (Version 2.4.18 or 2.4.20): $vulnerable"
printf "%s\n" "Total Version >= 2.4.0: $total24"
printf "Total: `jq 'select(.Agents[].Vendor=="Apache") | .IP' $tmpName | wc -l` \n"
printf '%s\n' '-----------------------'
rm $tmpName