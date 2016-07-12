#!/bin/bash
#Output filename: joomla_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
printf "`jq 'grep "Joomla" $1 | .CMS[] | select(.Vendor == "Joomla") | .Version' | sort | uniq -c | sort -nr` \n"
printf '%s\n' '-----------------------'