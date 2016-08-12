#!/bin/bash
#Output filename: embedded_server_AS_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '----------------------------------------------'
total=`grep -i 'mbedthis\|RomPager\|"Allegro Software RomPager"\|GoAhead' $1 | wc -l`

printf '%s\n' "------------Total entries: $total-------------"

grep -i 'mbedthis\|RomPager\|"Allegro Software RomPager"\|GoAhead' $1 | jq 'select(.Agents[].Vendor == "GoAhead" or .Agents[].Vendor == "mbedthis" or .Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .ASId' |  sort | uniq -c | sort -nr

printf '%s\n' '----------------------------------------------'