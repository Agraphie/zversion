#!/bin/bash
#Output filename: rompager_AS_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -i 'RomPager\|"Allegro Software RomPager"' $1  > $tmpName

printf '%s\n' '----------------------------------------------'
total=`jq 'select(.Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .IP' $tmpName | wc -l`

printf '%s\n' "------------Total entries: $total-------------"

jq 'select(.Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .ASId' $tmpName |  sort | uniq -c | sort -nr
rm $tmpName
printf '%s\n' '----------------------------------------------'