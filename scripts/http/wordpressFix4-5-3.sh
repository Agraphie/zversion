#!/bin/bash
#Output file name: wordpress_fix_4_5_3
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '------------------------------------------------------------'
#ips=($(grep "WordPress" $1 | jq 'select(.CMS[] | select(.Vendor=="WordPress" and .CanonicalVersion >= "0004000500030000" and .CanonicalVersion != "")) | .IP'))
printf "WordPress version >= 4.5.3:  `grep "WordPress" $1 | jq '.CMS[] | select(.Vendor=="WordPress" and .CMS[].CanonicalVersion == "0004000500030000" and .CMS[].CanonicalVersion != "") | .Vendor' | wc -l`"
printf "\n"
printf "WordPress total count: `grep "WordPress" $1 | jq '.CMS[] | select(.Vendor=="WordPress") | .Vendor' | wc -l` \n"
printf '%s\n' '-----------WordPress version 4.5.3 Top 10 ASN------------'
#asns=()
printf "`grep "WordPress" $1 | jq 'select(.CMS[].Vendor=="WordPress" and .CMS[].CanonicalVersion == "0004000500030000" and .CMS[].CanonicalVersion != "") | .ASId' | sort | uniq -c | sort -nr` \n"

#for i in "${ips[@]}"
#do
#    #remove quotes, this is necessary for jq to work
#    temp="${i%\"}"
#    temp="${temp#\"}"
#    asn=`grep -m 1 "$i" $1 |jq --arg ip $temp 'select(.IP == $ip) | .ASId'`
#    asns+=("$asn")
#done
#
#if [ ${#ips[@]} -gt 0 ]; then
#    echo "${asns[@]}" | tr ' ' '\n' | sort | uniq -c | sort -nr | head -n 10
#fi
printf '%s\n' '------------------------------------------------------------'
