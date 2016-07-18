#!/bin/bash
#Output filename: nginx_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------Nginx version distribution-------------'
printf "`grep "nginx" $1 | jq '.Agents[] | select(.Vendor == "nginx") | .Version' | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`grep "nginx" $1 | jq '.Agents[] | select(.Vendor == "nginx") | .Version' | wc -l`"
printf '%s\n' '-----------------------------------------------------'