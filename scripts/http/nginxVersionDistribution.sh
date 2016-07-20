#!/bin/bash
#Output filename: nginx_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------Nginx version distribution-------------'
printf "`grep "nginx" $1 | jq '.Agents[] | select(.Vendor == "nginx") | .Version' | sort | uniq -c | sort -nr` \n"
total1_11=`grep "nginx" $1 | jq '.Agents[] | select(.Vendor == "nginx" and .CanonicalVersion >= "0001001100000000" and .Version != "") | .CanonicalVersion' | wc -l`
total1_10=`grep "nginx" $1 | jq '.Agents[] | select(.Vendor == "nginx" and .Version != "" and .CanonicalVersion >= "0001001000000000"  and .CanonicalVersion < "0001001100000000") | .Version' | wc -l`
printf '\nTotal Version 1.10 <= Version <= 1.11: %s' "$total1_10"
printf '\nTotal Version >= 1.11: %s' "$total1_11"
printf  '\nTotal: %s\n' "`grep "nginx" $1 | jq '.Agents[] | select(.Vendor == "nginx") | .Version' | wc -l`"
printf '%s\n' '-----------------------------------------------------'