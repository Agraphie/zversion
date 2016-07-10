#!/bin/bash
#Output filename: drupal_version_distribution
printf $1
printf "`jq '.CMS[] | select(.Vendor == "Drupal") | .Version' $1 | sort | uniq -c | sort -nr`"