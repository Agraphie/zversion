#!/bin/bash
#Output file name: CMS_top_3_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
joomla=$(jq '.CMS[] | select(.Vendor=="Joomla") | .Vendor' $1 | wc -l)
wordpress=$(jq '.CMS[] | select(.Vendor=="WordPress") | .Vendor' $1 | wc -l)
drupal=$(jq '.CMS[] | select(.Vendor=="Drupal") | .Vendor' $1 | wc -l)
printf "WordPress: $wordpress"
printf "\n"
printf "Joomla: $joomla"
printf "\n"
printf "Drupal: $drupal"
printf "\n"
printf "Total (WordPress+Joomla+Drupal): $(( drupal+wordpress+joomla ))"
printf "\n"
printf '%s\n' '-----------------------'