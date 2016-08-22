#!/bin/bash
#Output filename: nginx_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "nginx" $1 > $tmpName

printf '%s\n' '-------------Nginx version distribution-------------'
printf "`jq '.Agents[] | select(.Vendor=="nginx") | .Version' $tmpName | sort | uniq -c | sort -nr` \n"
total0_1=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000100000000"  and .CanonicalVersion < "0000000200000000") | .Version' $tmpName | wc -l`
total0_2=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000200000000"  and .CanonicalVersion < "0000000300000000") | .Version' $tmpName | wc -l`
total0_3=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000300000000"  and .CanonicalVersion < "0000000400000000") | .Version' $tmpName | wc -l`
total0_4=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000400000000"  and .CanonicalVersion < "0000000500000000") | .Version' $tmpName | wc -l`
total0_5=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000500000000"  and .CanonicalVersion < "0000000600000000") | .Version' $tmpName | wc -l`
total0_6=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000600000000"  and .CanonicalVersion < "0000000700000000") | .Version' $tmpName | wc -l`
total0_7=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000700000000"  and .CanonicalVersion < "0000000800000000") | .Version' $tmpName | wc -l`
total0_8=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000800000000"  and .CanonicalVersion < "0000000900000000") | .Version' $tmpName | wc -l`
total0_9=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0000000900000000"  and .CanonicalVersion < "0001000000000000") | .Version' $tmpName | wc -l`
total1_0=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000000000000"  and .CanonicalVersion < "0001000100000000") | .Version' $tmpName | wc -l`
total1_1=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000100000000"  and .CanonicalVersion < "0001000200000000") | .Version' $tmpName | wc -l`
total1_2=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000200000000"  and .CanonicalVersion < "0001000300000000") | .Version' $tmpName | wc -l`
total1_3=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000300000000"  and .CanonicalVersion < "0001000400000000") | .Version' $tmpName | wc -l`
total1_4=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000400000000"  and .CanonicalVersion < "0001000500000000") | .Version' $tmpName | wc -l`
total1_5=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000500000000"  and .CanonicalVersion < "0001000600000000") | .Version' $tmpName | wc -l`
total1_6=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000600000000"  and .CanonicalVersion < "0001000700000000") | .Version' $tmpName | wc -l`
total1_7=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000700000000"  and .CanonicalVersion < "0001000800000000") | .Version' $tmpName | wc -l`
total1_8=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000800000000"  and .CanonicalVersion < "0001000900000000") | .Version' $tmpName | wc -l`
total1_9=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001000900000000"  and .CanonicalVersion < "0001001000000000") | .Version' $tmpName | wc -l`
total1_10=`jq '.Agents[] | select(.Vendor=="nginx" and .Version != "" and .CanonicalVersion >= "0001001000000000"  and .CanonicalVersion < "0001001100000000") | .Version' $tmpName | wc -l`
total1_11=`jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0001001100000000" and .CanonicalVersion < "0001001200000000" and .Version != "") | .Version' $tmpName | wc -l`

printf '\nTotal Version 0.1 <= Version < 0.2: %s' "$total0_1"
printf '\nTotal Version 0.2 <= Version < 0.3: %s' "$total0_2"
printf '\nTotal Version 0.3 <= Version < 0.4: %s' "$total0_3"
printf '\nTotal Version 0.4 <= Version < 0.5: %s' "$total0_4"
printf '\nTotal Version 0.5 <= Version < 0.6: %s' "$total0_5"
printf '\nTotal Version 0.6 <= Version < 0.7: %s' "$total0_6"
printf '\nTotal Version 0.7 <= Version < 0.8: %s' "$total0_7"
printf '\nTotal Version 0.8 <= Version < 0.9: %s' "$total0_8"
printf '\nTotal Version 0.9 <= Version < 1.0: %s' "$total0_9"
printf '\nTotal Version 1.0 <= Version < 1.1: %s' "$total1_0"
printf '\nTotal Version 1.1 <= Version < 1.2: %s' "$total1_1"
printf '\nTotal Version 1.2 <= Version < 1.3: %s' "$total1_2"
printf '\nTotal Version 1.3 <= Version < 1.4: %s' "$total1_3"
printf '\nTotal Version 1.4 <= Version < 1.5: %s' "$total1_4"
printf '\nTotal Version 1.5 <= Version < 1.6: %s' "$total1_5"
printf '\nTotal Version 1.6 <= Version < 1.7: %s' "$total1_6"
printf '\nTotal Version 1.7 <= Version < 1.8: %s' "$total1_7"
printf '\nTotal Version 1.8 <= Version < 1.9: %s' "$total1_8"
printf '\nTotal Version 1.9 <= Version < 1.10: %s' "$total1_9"
printf '\nTotal Version 1.10 <= Version < 1.11: %s' "$total1_10"
printf '\nTotal Version 1.11 <= Version < 1.12: %s' "$total1_11"
printf  '\nTotal: %s\n' "`jq 'select(.Agents[].Vendor=="nginx") | .IP' $tmpName | wc -l`"
printf '%s\n' '-----------------------------------------------------'
rm $tmpName