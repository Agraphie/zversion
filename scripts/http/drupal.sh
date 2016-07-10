#!/bin/bash
#Output filename: drupal_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
printf "`jq '.CMS[] | select(.Vendor == "Drupal") | .Version' $1 | sort | uniq -c | sort -nr` \n"
printf '%s\n' '-----------------------'