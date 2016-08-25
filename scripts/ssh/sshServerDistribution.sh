#!/bin/bash
#Output filename: major_server_vendor_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '----------------------------------------------'
majorServerVendors=(
    "OpenSSH"
	"RomSShell"
	"dropbear"
	"sftp"
	"CoreFTP"
	"AppGateSSH"
	"SSH Tectia Server"
	"sshlib"
    ""
	)

total=`wc -l < $1`
totalNoErrors=`jq 'select(.Error == "") | .IP' $1 | wc -l`

printf '%s\n' "------------Total entries: $total ($totalNoErrors no error)-------------"

for i in "${majorServerVendors[@]}"
do
    vendorCount=0
    vendorCount=`grep "$i" $1 | jq --arg vendor "$i" 'select(.Vendor == $vendor) | .Vendor' |  wc -l`

    printf '%s\n' "$i: $vendorCount ($(awk "BEGIN {printf \"%.2f\n\", 100/$total*$vendorCount}")% of total)"
done
printf '%s\n' '-------------------------------------------------------------------------'