#!/bin/bash
#Output file name: wordpress_fix_4_5_3
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
ips=($(jq ' select(.CMS[] | select(.Vendor=="WordPress" and .CanonicalVersion >= "0004000500030000" and .CanonicalVersion != "")) | .IP' $1))
printf "WordPress version >= 4.5.3: ${#ips[@]}"
printf "\n"
printf "WordPress total count: `jq '.CMS[] | select(.Vendor=="WordPress") | .Vendor' $1 | wc -l` \n"
printf '%s\n' '-----------WordPress version >= 4.5.3 Top 10 ASN------------'
asns=()

for i in "${ips[@]}"
do
    #remove quotes, this is necessary for jq to work
    temp="${i%\"}"
    temp="${temp#\"}"
    asn=`jq --arg ip $temp 'select(.IP == $ip) | .ASId' $1 | head -n 1`
    asns+=("$asn")
done

echo "${asns[@]}" | tr ' ' '\n' | sort | uniq -c | sort -nr | head -n 10

printf '%s\n' '------------------------------------------------------------'
