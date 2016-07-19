#!/bin/bash
#Output filename: major_server_vendor_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '----------------------------------------------'
majorServerVendors=(
    "Microsoft-IIS"
	"Apache"
	"nginx"
	"lighttpd"
	"ATS"
	"BOA"
	"Allegro Software RomPager"
	"AllegroServe"
	"squid"
	"tengine"
	"jetty"
	"RomPager"
	"mini_httpd"
	"micro_httpd"
	"AOL Server"
	"Abyss"
	"Agranat-EmWeb"
	"Microsoft-HTTPAPI"
	"CherryPy"
	"Cherokee"
	"CommuniGatePro"
	"EdgePrism"
	"Flywheel"
	"GoAhead"
	"IdeaWebServer"
	"Indy"
	"mbedthis"
	"PRTG"
	"Kangle"
	"thttpd")

total=`wc -l < $1`

printf '%s\n' "------------Total entries: $total-------------"

for i in "${majorServerVendors[@]}"
do
    vendorCount=0
    vendorCount=`grep "$i" $1 | jq --arg vendor "$i" '.Agents[] | select(.Vendor == $vendor) | .Vendor' |  wc -l`

    printf '%s\n' "$i: $vendorCount"
done
printf '%s\n' '----------------------------------------------'