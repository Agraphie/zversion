#!/bin/bash
#Output filename: apache_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------Apache version distribution-------------'
printf "`grep "Apache" $1 | jq '.Agents[] | select(.Vendor == "Apache") | .Version' | sort | uniq -c | sort -nr` \n"
total24=`grep "Apache" $1 | jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion >= "0002000400000000" and .Version != "") | .Vendor' | wc -l`
total22=`grep "Apache" $1 | jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion <= "0002000400000000" and .CanonicalVersion >= "0002000200000000" and .Version != "") | .Vendor' | wc -l`
printf '\nTotal Version 2.2 <= Version <= 2.4: %s' "$total22"
printf '\nTotal Version >= 2.4: %s' "$total24"
printf  '\nTotal: %s\n' "`grep "Apache" $1 | jq '.Agents[] | select(.Vendor == "Apache") | .Version' | wc -l`"
printf '%s\n' '-----------------------------------------------------'