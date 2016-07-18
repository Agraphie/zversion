#!/bin/bash
#Output filename: server_distribution_for_continents
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
continents=( "EU" "AS" "AF" "AN" "SA" "NA" "OC" )

for i in "${continents[@]}"
do
    #remove quotes, this is necessary for jq to work
    temp="${i%\"}"
    temp="${temp#\"}"
    top3=`grep "$temp" $1 | jq --arg cont "$temp" 'select(.GeoData.Continent == $cont) | .Agents[].Vendor' |  sort | uniq -c | sort -nr | head -n 3`
    if  [[ !  -z  $top3  ]]; then
        printf '%s\n' "---------- $temp -------------"
        printf '%s\n' "$top3"
        printf "\n"
    fi
done
printf '%s\n' '-----------------------'