#!/bin/bash
#Output filename: rompager_AS_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '----------------------------------------------'
total=`grep -i 'RomPager\|"Allegro Software RomPager"' $1 | wc -l`

printf '%s\n' "------------Total entries: $total-------------"

grep -i 'RomPager\|"Allegro Software RomPager"' $1 | jq 'select(.Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .ASId' |  sort | uniq -c | sort -nr

printf '%s\n' '----------------------------------------------'