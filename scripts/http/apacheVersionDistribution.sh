#!/bin/bash
#Output filename: apache_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "Apache" $1 > $tmpName

printf '%s\n' '-------------Apache version distribution-------------'
printf "`jq '.Agents[] | select(.Vendor=="Apache") | .Version' $tmpName | sort | uniq -c | sort -nr` \n"
total24=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion >= "0002000400000000" and .CanonicalVersion <= "0002000400230000" and  .Version != "") | .Vendor' $tmpName | wc -l`
total22=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion < "0002000400000000" and .CanonicalVersion <= "0002000200310000" and .CanonicalVersion >= "0002000200000000" and .Version != "") | .Vendor' $tmpName | wc -l`
total20=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion < "0002000200000000" and .CanonicalVersion <= "0002000000650000" and .CanonicalVersion >= "0002000000000000" and .Version != "") | .Vendor' $tmpName | wc -l`
total13=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion < "0002000000000000" and .CanonicalVersion <= "0001000300420000" and .CanonicalVersion >= "0001000300000000" and .Version != "") | .Vendor' $tmpName | wc -l`
totalVersion=`jq '.Agents[] | select(.Vendor=="Apache" and .Version != "") | .Vendor' $tmpName | wc -l`
total=`jq 'select(.Agents[].Vendor=="Apache") | .IP' $tmpName | wc -l`
totalUnknownVersion=$((total-totalVersion))

rm $tmpName
printf '\nTotal Version 1.3 <= Version < 2.0: %s' "$total13 "
printf '\nTotal Version 2.0 <= Version < 2.2: %s' "$total20"
printf '\nTotal Version 2.2 <= Version < 2.4: %s' "$total22"
printf '\nTotal Version >= 2.4: %s' "$total24"

printf '\n%s' '-----------------------'
printf '\nTotal known version: %s' "$totalVersion ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$totalVersion}")%)"
printf '\nTotal unknown version: %s' "$totalUnknownVersion ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$totalUnknownVersion}")%)"

printf  '\nTotal: %s\n' "$total"
printf '%s\n' '-----------------------------------------------------'