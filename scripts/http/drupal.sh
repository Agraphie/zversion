#!/bin/bash
printf $1
printf "\n Drupal \n"
printf "`jq '.CMS[] | select(.Vendor == "Drupal") | .Version' $1 | sort | uniq -c | sort -nr`"
printf "\n done \n"
