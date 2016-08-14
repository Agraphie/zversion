#!/bin/bash
#Output filename: apache1_3_and_2_0_AS_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
total20=`grep "Apache" $1 | jq 'select(.Agents[].Vendor=="Apache" and .Agents[].CanonicalVersion < "0002000200000000" and .Agents[].CanonicalVersion <= "0002000000650000" and .Agents[].CanonicalVersion >= "0002000000000000" and .Agents[].Version != "") | .ASId' | sort | uniq -c | sort -nr`
total13=`grep "Apache" $1 | jq 'select(.Agents[].Vendor=="Apache" and .Agents[].CanonicalVersion < "0002000000000000" and .Agents[].CanonicalVersion <= "0001000300420000" and .Agents[].CanonicalVersion >= "0001000300000000" and .Agents[].Version != "") | .ASId' | sort | uniq -c | sort -nr`
printf '%s\n' '-------------Apache 1.3 version distribution-------------'
printf '%s\n' "$total13"

printf '%s\n' '-------------Apache 2.0 version distribution-------------'
printf '%s\n' "$total20"
printf '%s\n' '-----------------------------------------------------'