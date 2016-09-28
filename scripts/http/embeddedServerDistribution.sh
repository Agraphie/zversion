#!/bin/bash
#Output filename: embedded_server_AS_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -i 'mbedthis\|RomPager\|"Allegro Software RomPager"\|GoAhead' $1 > $tmpName
printf '%s\n' '----------------------------------------------'
total=`wc -l < $tmpName`

printf '%s\n' "------------Total entries: $total-------------"

grep -i 'mbedthis\|RomPager\|"Allegro Software RomPager"\|GoAhead' $tmpName | jq 'select(.Agents[].Vendor == "GoAhead" or .Agents[].Vendor == "mbedthis" or .Agents[].Vendor == "RomPager" or .Agents[].Vendor == "Allegro Software RomPager") | .ASId' |  sort | uniq -c | sort -nr

printf '%s\n' '----------------------------------------------'
rm $tmpName