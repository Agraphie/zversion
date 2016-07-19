#!/bin/bash
#Output filename: cms_top_3_for_asn
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
asns=($(jq '.ASId' $1 | sort | uniq -c | sort -nr | awk '$1 > 100  {print $2 $3}'))
for i in "${asns[@]}"
do
    #remove quotes, this is necessary for jq to work
    temp="${i%\"}"
    temp="${temp#\"}"
    top3=`grep "$i" $1 | jq '.CMS[].Vendor' |  sort | uniq -c | sort -nr | head -n 3`
    if  [[ !  -z  $top3  ]]; then
        asnName=`grep -m 1 "$i" $1 | jq --arg asn $temp 'select(.ASId==$asn) | .ASOwner'`
        printf '%s\n' "---------- $temp ($asnName) -------------"
        printf '%s\n' "$top3"
        printf "\n"
    fi
done
printf '%s\n' '-----------------------'