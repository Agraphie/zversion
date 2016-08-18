#!/bin/bash
#Output filename: nginx_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "nginx" $1 > $tmpName

printf '%s\n' '-------------Nginx version distribution-------------'
printf "`jq '.Agents[] | select(.Vendor=="nginx") | .Version' $tmpName | sort | uniq -c | sort -nr` \n"
total1_11=`jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0001001100000000" and .Version != "") | .CanonicalVersion' $tmpName | wc -l`
total1_10=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001001000000000"  and .CanonicalVersion < "0001001100000000") | .Version' $tmpName | wc -l`
printf '\nTotal Version 1.10 <= Version < 1.11: %s' "$total1_10"
printf '\nTotal Version >= 1.11: %s' "$total1_11"
printf  '\nTotal: %s\n' "`jq 'select(.Agents[].Vendor=="nginx") | .IP' $tmpName | wc -l`"
printf '%s\n' '-----------------------------------------------------'
rm $tmpName