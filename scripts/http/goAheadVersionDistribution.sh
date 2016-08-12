#!/bin/bash
#Output filename: goahead_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------RomPager version distribution-------------'
printf "`grep "GoAhead" $1 | jq '.Agents[] | select(.Vendor == "GoAhead") | .Version' | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`grep "GoAhead" $1 | jq '.Agents[] | select(.Vendor == "GoAhead") | .Version' | wc -l`"
printf '%s\n' '-----------------------------------------------------'