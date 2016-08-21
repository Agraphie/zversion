#!/bin/bash
#Output filename: microsoft_iis_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -i "Microsoft" $1 > $tmpName

printf '%s\n' '-------------Microsoft-IIS version distribution-------------'
printf "`jq '.Agents[] | select(.Vendor=="Microsoft-IIS") | .Version' $tmpName | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`jq 'select(.Agents[].Vendor=="Microsoft-IIS") | .IP' $tmpName | wc -l`"
printf '%s\n' '-----------------------------------------------------'
rm $tmpName