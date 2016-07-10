#!/bin/bash
#Output filename: wordpress_version_distribution
printf $1
printf "`jq '.CMS[] | select(.Vendor == "WordPress") | .Version' $1 | sort | uniq -c | sort -nr`"