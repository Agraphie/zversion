#!/bin/bash
#Output filename: server_AS_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '----------------------------------------------'

total=`wc -l < $1`
dir=$(dirname "$(readlink -f "$0")")

printf '%s\n' "------------Total entries: $total-------------"
$dir/../jq '.ASId' $1 | sort | uniq -c | sort -nr

printf '%s\n' '-------------------------------------------------------------------------'