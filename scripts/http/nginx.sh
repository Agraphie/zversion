#!/bin/bash
#Output file name: nginx_major_vulnerabilities_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
printf '%s\n' '-----------------------'
printf "'CVE-2014-0133 (1.3.15 - 1.5.11)'; `jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0001000300150000" and .CanonicalVersion <= "0001000500110000" and .CanonicalVersion != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2014-0088 (1.5.10)'; `jq '.Agents[] | select(.Agent=="Vendor" and .CanonicalVersion == "0001000500100000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2013-2028 (1.3.9 - 1.4.0)'; `jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0001000300090000" and .CanonicalVersion <= "0001000400000000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2012-2089 (1.1.3 - 1.1.18, 1.0.7 - 1.0.14)'; `jq '.Agents[] | select(.Vendor=="nginx" and ((.CanonicalVersion >= "0001000100030000" and .CanonicalVersion <= "0001000100180000") or (.CanonicalVersion >= "0001000000070000" and .CanonicalVersion <= "0001000000140000")) and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2012-1180 (0.1.0 - 1.1.16)'; `jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0000000100000000" and .CanonicalVersion <= "0001000100160000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2009-3555 (0.1.0 - 0.8.22)'; `jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0000000100000000" and .CanonicalVersion <= "0000000800220000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2009-2629 (0.1.0 - 0.8.14)'; `jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0000000100000000" and .CanonicalVersion <= "0000000800140000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"
printf "'CVE-2009-3896 (0.1.0 - 0.8.13)'; `jq '.Agents[] | select(.Vendor=="nginx" and .CanonicalVersion >= "0000000100000000" and .CanonicalVersion <= "0000000800130000" and .Version != "") | .Version' $1 | wc -l`"
printf "\n"

printf "Total: `jq '.Agents[] | select(.Vendor=="nginx") | .Vendor' $1 | wc -l` \n"
printf '%s\n' '-----------------------'