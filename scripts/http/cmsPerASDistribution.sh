#!/bin/bash
#Output filename: cms_top_3_for_asn
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
asns=($(jq '.ASId' $1 | sort | uniq))
for i in "${asns[@]}"
do
    #remove quotes, this is necessary for jq to work
    temp="${i%\"}"
    temp="${temp#\"}"
    top3=`jq --arg asn $temp 'select(.ASId == $asn) | .CMS[].Vendor' $1 |  sort | uniq -c | sort -nr | head -n 3`
    if  [[ !  -z  $top3  ]]; then
    printf '%s\n' "----------$temp-------------"
    printf '%s\n' "$top3"
    printf "\n"
    fi
done
printf '%s\n' '-----------------------'