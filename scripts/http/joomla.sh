#!/bin/bash
#Output filename: joomla_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------Joomla version distribution-------------'
printf "`grep "Joomla" $1 | jq '.CMS[] | select(.Vendor == "Joomla") | .Version' | sort | uniq -c | sort -nr` \n"
printf '%s\n' '-----------------------------------------------------'