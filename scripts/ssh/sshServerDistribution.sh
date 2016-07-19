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
	)

total=`wc -l < $1`

printf '%s\n' "------------Total entries: $total-------------"

for i in "${majorServerVendors[@]}"
do
    vendorCount=0
    vendorCount=`grep "$i" $1 | jq --arg vendor "$i" 'select(.Vendor == $vendor) | .Vendor' |  wc -l`

    printf '%s\n' "$i: $vendorCount"
done
printf '%s\n' '----------------------------------------------'