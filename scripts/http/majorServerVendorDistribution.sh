#!/bin/bash
#Output filename: major_server_vendor_distribution_for_asn
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
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

for i in "${majorServerVendors[@]}"
do
    total=`grep $i $1 |  wc -l`
    printf '%s\n' "------Vendor: $i ($total in total)-------"

    if [[ $total > 0 ]]; then
        top3=`grep $i | jq --arg vendor $i 'select(.Agents[].Vendor == $vendor) | .ASId' |  sort | uniq -c | sort -nr | head -n 10`
        printf '%s\n' "$top3"
        printf "\n"
    fi
done
printf '%s\n' '-----------------------'