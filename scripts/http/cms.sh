#!/bin/bash
#Output file name: CMS_top_3_distribution
printf $1
printf "\n"
printf "WordPress: `jq '.CMS[] | select(.Vendor=="WordPress") | .Vendor' $1 | wc -l`"
printf "\n"
printf "Joomla: `jq '.CMS[] | select(.Vendor=="Joomla") | .Vendor' $1 | wc -l`"
printf "\n"
printf "Drupal: `jq '.CMS[] | select(.Vendor=="Drupal") | .Vendor' $1 | wc -l`"
printf "\n"
printf "Total: `jq '.CMS[] | select(.Vendor !="") | .Vendor' $1 | wc -l`"
printf "\n"
