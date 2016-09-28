#!/bin/bash
#Output filename: error_classification
printf "Script name: $0\n"
printf "Input file: $1\n"
#create tmp file
tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -v '"Error":""' $1 > $tmpName
total=$(wc -l < $tmpName)

printf '%s\n' "-----------Total errors: $total------------"
jq 'select(.Error != "") | .Error' $tmpName | sort | uniq -c | sort -nr
printf '%s\n' '-------------------------------------------------------------------------'
rm $tmpName
