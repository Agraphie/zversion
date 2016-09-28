#!/bin/bash
#Output file name: dropbear_cve_2012-0920
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '----------------------------------------------'
printf "0.52 <= Dropbear Version <= 2011.54: `grep "dropbear" $1 |  jq 'select(.Vendor =="dropbear" and .CanonicalVersion <= "2011005400000000" and .CanonicalVersion >= "0000005200000000" and .SoftwareVersion != "") | .SoftwareVersion' | wc -l`"
printf "\n"
printf "Total dropbear: `grep "dropbear" $1 | jq 'select(.Vendor =="dropbear") | .Vendor' | wc -l`\n"
printf '%s\n' '----------------------------------------------'