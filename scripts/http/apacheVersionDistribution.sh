#!/bin/bash
#Output filename: apache_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------Apache version distribution-------------'
printf "`grep "Apache" $1 | jq '.Agents[] | select(.Vendor == "Apache") | .Version' | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`grep "Apache" $1 | jq '.Agents[] | select(.Vendor == "Apache") | .Version' | wc -l`"
printf '%s\n' '-----------------------------------------------------'