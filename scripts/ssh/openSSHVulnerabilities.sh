#!/bin/bash
#Output file name: openssh_vulnerabilities_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep "OpenSSH" $1 > $tmpName

printf '%s\n' '----------------------------------------------'
CVE_2016_6515=$(grep -v '5ubuntu1.10"\|2ubuntu2.8"\|4ubuntu2.1"\|4+deb7u6"' $tmpName | jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion != "" and .CanonicalVersion <= "0007000300000000" and .Comments != "5ubuntu1.8") | .IP' | wc -l)

CVE_2015_8325=$(grep -v '4+deb7u4"\|4+deb7u6"\|5+deb8u2"\|5+deb8u3"\|5ubuntu1.9"\|2ubuntu2.7"\|2ubuntu0.2"' $tmpName | jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion == "7.2p2") | .IP' | wc -l)

CVE_2015_6565=$(grep -v '4+deb7u4"\|4+deb7u6"\|5+deb8u2"\|5+deb8u3"' $tmpName | jq 'select(.Vendor=="OpenSSH" and (.SoftwareVersion == "6.8" or .SoftwareVersion == "6.9")) | .IP' | wc -l)

CVE_2015_6564=$(jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion == "6.9") | .IP' $tmpName | wc -l)

CVE_2015_5600=$(grep -v '5ubuntu1.6"\|2ubuntu2.2"\|5ubuntu1.2"\|6ubuntu1"' $tmpName | jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion == "6.9") | .IP' | wc -l)

CVE_2014_1692=$(grep -v '4+deb7u4"\|4+deb7u6"\|5+deb8u2"\|5+deb8u3"' $tmpName | jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion != "" and .CanonicalVersion <= "0006000400000000") | .IP' | wc -l)

CVE_2013_4548=$(grep -v '6ubuntu1"' $tmpName | jq 'select(.Vendor=="OpenSSH" and (.SoftwareVersion == "6.2" or .SoftwareVersion == "6.3")) | .IP' | wc -l)

CVE_2010_4478=$(jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion != "" and .CanonicalVersion <= "0005000600000000") | .IP' $tmpName | wc -l)
CVE_2009_2904=$(jq 'select(.Vendor=="OpenSSH" and (.SoftwareVersion == "4.3" or .SoftwareVersion == "4.8")) | .IP' $tmpName | wc -l)
CVE_2008_3234=$(jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion == "4.0") | .IP' $tmpName | wc -l)



printf "%s\n" "CVE-2016-6515: $CVE_2016_6515"
printf "%s\n" "CVE-2015-8325: $CVE_2015_8325"
printf "%s\n" "CVE-2015-6565: $CVE_2015_6565"
printf "%s\n" "CVE-2015-6564: $CVE_2015_6564"
printf "%s\n" "CVE-2015-5600: $CVE_2015_5600"

printf "%s\n" "CVE-2014-1692: $CVE_2014_1692"
printf "%s\n" "CVE-2013-4548: $CVE_2013_4548"
printf "%s\n" "CVE-2010-4478: $CVE_2010_4478"
printf "%s\n" "CVE-2009-2904: $CVE_2009_2904"
printf "%s\n" "CVE-2008-3234: $CVE_2008_3234"


printf "Total: `jq 'select(.Vendor=="OpenSSH") | .IP'  $tmpName | wc -l` \n"
printf '%s\n' '----------------------------------------------'
rm $tmpName