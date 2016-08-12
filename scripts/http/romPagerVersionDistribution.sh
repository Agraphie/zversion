#!/bin/bash
#Output filename: rompager_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------RomPager version distribution-------------'
printf "`grep "RomPager" $1 | jq '.Agents[] | select(.Vendor == "RomPager") | .Version' | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`grep "RomPager" $1 | jq '.Agents[] | select(.Vendor == "RomPager") | .Version' | wc -l`"
printf '%s\n' '-----------------------------------------------------'