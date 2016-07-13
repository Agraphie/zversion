#!/bin/bash
#Output filename: server_vendor_top_3_for_asn
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '------------Server Distribution for ASNS appearing at least 100 times-------------'
asns=($(grep -v "Agents\[\]" $1| jq '.ASId' | sort | uniq -c | sort -nr | awk '$1 >= 100  {print $2 $3}'))
printf "#ASNS: %s\n" "${#asns[@]}"

for i in "${asns[@]}"
do
    #remove quotes, this is necessary for jq to work
    temp="${i%\"}"
    temp="${temp#\"}"
    top3=`grep "$i" $1 |jq --arg asn $temp 'select(.ASId == $asn) | .Agents[].Vendor' |  sort | uniq -c | sort -nr | head -n 5`
    if  [[ !  -z  $top3  ]]; then
        asnName=`grep -m 1 "$i" $1 | jq --arg asn "$temp" 'select(.ASId==$asn) | .ASOwner'`
        printf '%s\n' "---------- $temp ($asnName) -------------"
        printf '%s\n' "$top3"
        printf "\n"
    fi
done
printf '%s\n' '-----------------------------------------------------------------------------------'