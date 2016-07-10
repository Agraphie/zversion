#!/bin/bash
printf $1
printf "\n Joomla \n"
printf "`jq '.CMS[] | select(.Vendor == "Joomla") | .Version' $1 | sort | uniq -c | sort -nr`"
printf "\n Done \n"
