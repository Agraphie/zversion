#!/bin/bash
#Output filename: dropbear_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------dropbear version distribution-------------'
printf "`grep "dropbear" $1 | jq 'select(.Vendor == "dropbear") | .SoftwareVersion' | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`grep "dropbear" $1 | jq 'select(.Vendor == "dropbear") | .SoftwareVersion' | wc -l`"
printf '%s\n' '-----------------------------------------------------'