#!/bin/bash
printf $1
printf "\n WordPress \n"
printf "`jq '.CMS[] | select(.Vendor == "WordPress") | .Version' $1 | sort | uniq -c | sort -nr`"
printf "\n done \n"
