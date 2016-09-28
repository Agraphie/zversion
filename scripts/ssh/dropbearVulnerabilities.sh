#!/bin/bash
#Output file name: dropbear_vulnerabilities_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "dropbear" $1 > $tmpName

printf '%s\n' '----------------------------------------------'
CVE_2013_4434AndCVE_2013_4421=$(jq 'select(.Vendor=="dropbear" and . Version != "" and .CanonicalVersion >= "0000002800000000" and .CanonicalVersion <= "2013005800000000") | .IP' $tmpName | wc -l)
CVE_2012_0920=$(jq 'select(.Vendor=="dropbear" and . Version != "" and .CanonicalVersion >= "0000005200000000" and .CanonicalVersion <= "2012005400000000") | .IP' $tmpName | wc -l)
CVE_2007_1099=$(jq 'select(.Vendor=="dropbear" and . Version != "" and .CanonicalVersion >= "0000004000000000" and .CanonicalVersion <= "0000004800010000") | .IP' $tmpName | wc -l)
CVE_2006_1206=$(jq 'select(.Vendor=="dropbear" and . Version != "" and .CanonicalVersion >= "0000002800000000" and .CanonicalVersion <= "0000004700000000") | .IP' $tmpName | wc -l)
CVE_2005_4178=$(jq 'select(.Vendor=="dropbear" and . Version != "" and .CanonicalVersion >= "0000002800000000" and .CanonicalVersion <= "0000004600000000") | .IP' $tmpName | wc -l)
CVE_2004_2486=$(jq 'select(.Vendor=="dropbear" and . Version != "" and .CanonicalVersion >= "0000002800000000" and .CanonicalVersion <= "0000004200000000") | .IP' $tmpName | wc -l)


printf "%s\n" "CVE-2013-4434 and CVE-2013-4421: $CVE_2013_4434AndCVE_2013_4421"
printf "%s\n" "CVE-2012-0920: $CVE_2012_0920"
printf "%s\n" "CVE-2007-1099: $CVE_2007_1099"
printf "%s\n" "CVE-2006-1206: $CVE_2006_1206"
printf "%s\n" "CVE-2005-4178: $CVE_2005_4178"
printf "%s\n" "CVE-2004-2486: $CVE_2004_2486"


printf "Total: `jq 'select(.Vendor=="dropbear") | .IP'  $tmpName | wc -l` \n"
printf '%s\n' '----------------------------------------------'
rm $tmpName