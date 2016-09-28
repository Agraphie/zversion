#!/bin/bash
#Output file name: apache_important_vulnerabilities_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "Apache" $1 > $tmpName

printf '%s\n' '------------------Apache 2.4 vulnerable----------------------------'
CVE_2016_4979=`grep 'Apache","Version":"2.4.18\|Apache","Version":"2.4.20' $tmpName | jq '.Agents[] | select(.Vendor=="Apache" and (.CanonicalVersion == "0002000400180000" or .CanonicalVersion == "0002000400200000") and .Version != "") | .Version' | wc -l`
CVE_2014_0231AndCVE_2014_3523=`jq '.Agents[] | select(.Vendor=="Apache" and (.Version == "2.4.9" or .Version =="2.4.7" or .Version =="2.4.6" or .Version == "2.4.4" or .Version =="2.4.3" or .Version == "2.4.2" or .Version == "2.4.1") and .Version != "") | .Version' $tmpName  | wc -l`
CVE_2012_3502=`jq '.Agents[] | select(.Vendor=="Apache" and (.Version == "2.4.2" or .Version ==" 2.4.1") and .Version != "") | .Version' $tmpName  | wc -l`

printf "%s\n" "CVE-2016-4979: $CVE_2016_4979"
printf "%s\n" "CVE-2014-0231 and CVE-2014-3523: $CVE_2014_0231AndCVE_2014_3523"
printf "%s\n" "CVE-2012-3502: $CVE_2012_3502"

printf '%s\n' '------------------Apache 2.2 vulnerable----------------------------'
CVE_2014_0231=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion <= "0002000200270000" and .CanonicalVersion >= "0002000200000000" and .Version != "2.2.7" and .Version != "2.2.1" and .Version != "") | .Version' $tmpName | wc -l`
CVE_2011_3192=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion <= "0002000200190000" and .CanonicalVersion >= "0002000200000000" and .Version != "2.2.7" and .Version != "2.2.1" and .Version != "") | .Version' $tmpName | wc -l`
CVE_2010_2068=`jq '.Agents[] | select(.Vendor=="Apache" and ((.CanonicalVersion <= "0002000200150000" and .CanonicalVersion >= "0002000200090000") or .Version == "2.3.5" or .Version == "2.3.4") and .Version != "") | .Version' $tmpName | wc -l`
CVE_2010_0425=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion <= "0002000200140000" and .CanonicalVersion >= "0002000200000000" and .Version != "2.2.7" and .Version != "2.2.1" and .Version != "") | .Version' $tmpName | wc -l`
CVE_2009_1890=`jq '.Agents[] | select(.Vendor=="Apache" and .CanonicalVersion <= "0002000200110000" and .CanonicalVersion >= "0002000200000000" and .Version != "2.2.7" and .Version != "2.2.1" and .Version != "") | .Version' $tmpName | wc -l`
CVE_2009_1191=`grep 'Apache","Version":"2.2.11' $tmpName | jq '.Agents[] | select(.Vendor=="Apache" and .Version == "2.2.11") | .Version'  | wc -l`
CVE_2010_2791=`grep 'Apache","Version":"2.2.9' $tmpName | jq '.Agents[] | select(.Vendor=="Apache" and .Version == "2.2.9") | .Version' | wc -l`
CVE_2006_3747=`jq '.Agents[] | select(.Vendor=="Apache" and (.Version == "2.2.0" or .Version == "2.2.2")) | .Version' $tmpName | wc -l`

printf "%s\n" "CVE-2014-0231: $CVE_2014_0231"
printf "%s\n" "CVE-2011-3192: $CVE_2011_3192"
printf "%s\n" "CVE-2010-2068: $CVE_2010_2068"
printf "%s\n" "CVE-2010-0425: $CVE_2010_0425"
printf "%s\n" "CVE-2009-1890: $CVE_2009_1890"
printf "%s\n" "CVE-2009-1191: $CVE_2009_1191"
printf "%s\n" "CVE-2010-2791: $CVE_2010_2791"
printf "%s\n" "CVE-2006-3747: $CVE_2006_3747"

printf "Total: `jq 'select(.Agents[].Vendor=="Apache") | .IP'  $tmpName | wc -l` \n"
printf '%s\n' '----------------------------------------------'
rm $tmpName