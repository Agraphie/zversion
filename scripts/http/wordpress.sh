#!/bin/bash
#Output filename: wordpress_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------WordPress version distribution-------------'
printf "`grep "WordPress" $1 | jq '.CMS[] | select(.Vendor == "WordPress") | .Version' | sort | uniq -c | sort -nr`\n"
printf '%s\n' '--------------------------------------------------------'