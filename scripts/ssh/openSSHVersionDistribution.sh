#!/bin/bash
#Output filename: openssh_version_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-------------OpenSSH version distribution-------------'
printf "`grep "OpenSSH" $1 | jq 'select(.Vendor == "OpenSSH") | .SoftwareVersion' | sort | uniq -c | sort -nr` \n"
printf  '\nTotal: %s\n' "`grep "OpenSSH" $1 | jq 'select(.Vendor == "OpenSSH") | .SoftwareVersion' | wc -l`"
printf '%s\n' '-----------------------------------------------------'