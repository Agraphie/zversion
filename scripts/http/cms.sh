#!/bin/bash
#Output file name: CMS_top_3_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
joomla=$(grep "Joomla" $1 | jq '.CMS[] | select(.Vendor=="Joomla") | .Vendor' | wc -l)
wordpress=$(grep "WordPress" $1 | jq '.CMS[] | select(.Vendor=="WordPress") | .Vendor'  | wc -l)
drupal=$(grep "Drupal" $1 | jq '.CMS[] | select(.Vendor=="Drupal") | .Vendor' | wc -l)
printf "WordPress: $wordpress"
printf "\n"
printf "Joomla: $joomla"
printf "\n"
printf "Drupal: $drupal"
printf "\n"
printf "Total (WordPress+Joomla+Drupal): $(( drupal+wordpress+joomla ))"
printf "\n"
printf '%s\n' '-----------------------'