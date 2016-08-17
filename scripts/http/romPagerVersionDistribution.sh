#!/bin/bash
#Output filename: rompager_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -i 'RomPager\|"Allegro Software RomPager"' $1  > $tmpName

printf '%s\n' '-------------RomPager version distribution-------------'
printf "`jq 'select(.Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .Agents[].Version' $tmpName | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`jq 'select(.Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .IP' $tmpName | wc -l`"
printf '%s\n' '-----------------------------------------------------'
rm $tmpName