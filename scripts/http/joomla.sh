#!/bin/bash
#Output filename: joomla_version_distribution
printf $1
printf "`jq '.CMS[] | select(.Vendor == "Joomla") | .Version' $1 | sort | uniq -c | sort -nr`"