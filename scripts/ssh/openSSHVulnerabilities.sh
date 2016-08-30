#!/bin/bash
#Output file name: openssh_vulnerabilities_distribution
printf "Script name: $0\n"
printf "Input file: $1\n"
dir=$(dirname "$(readlink -f "$0")")

tmpName=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1).tmp
grep -F "OpenSSH" $1 > $tmpName

printf '%s\n' '----------------------------------------------'
CVE_2016_6515=$($dir/../jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion != "" and .CanonicalVersion <= "0007000300000000"
and select(.Comments | endswith("5ubuntu1.10") == false and endswith("2ubuntu2.8") == false and
endswith("4ubuntu2.1") == false and endswith("4+deb7u6") == false)) | .IP' $tmpName  | wc -l)

CVE_2015_8325=$($dir/../jq 'select(.Vendor=="OpenSSH" and .CanonicalVersion <= "0007000200000000" and select(.Comments | endswith("4+deb7u4") == false and endswith("4+deb7u6") == false and endswith("5+deb8u2") == false and endswith("5+deb8u3") == false and endswith("5ubuntu1.9") == false and endswith("2ubuntu2.7") == false and endswith("2ubuntu0.2") == false)) | .IP' $tmpName | wc -l)

CVE_2015_6565=$($dir/../jq 'select(.Vendor=="OpenSSH" and (.SoftwareVersion == "6.8" or .SoftwareVersion == "6.9") and select(.Comments | endswith("4+deb7u4") == false and endswith("4+deb7u6") == false and endswith("5+deb8u2") == false and endswith("5+deb8u3") == false)) | .IP' $tmpName | wc -l)

CVE_2015_6564=$($dir/../jq 'select(.Vendor=="OpenSSH" and .CanonicalVersion <= "0006000900000000" and select(.SoftwareVersion | endswith("p") == false)) | .IP' $tmpName | wc -l)

CVE_2015_5600=$($dir/../jq 'select(.Vendor=="OpenSSH" and .CanonicalVersion <= "0006000900000000"
and select(.Comments |  endswith("5ubuntu1.6") == false and endswith("2ubuntu2.2") == false and endswith("5ubuntu1.2") == false and endswith("6ubuntu1") == false)) | .IP' $tmpName  | wc -l)

CVE_2014_1692=$($dir/../jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion != "" and .CanonicalVersion <= "0006000400000000" and select(.Comments | endswith("4+deb7u4") == false and endswith("4+deb7u6") == false and endswith("5+deb8u2") == false and endswith("5+deb8u3") == false)) | .IP' $tmpName | wc -l)

CVE_2013_4548=$($dir/../jq 'select(.Vendor=="OpenSSH" and (.SoftwareVersion == "6.2" or .SoftwareVersion == "6.3") and
select(.Comments | endswith("6ubuntu1") == false)
) | .IP' $tmpName | wc -l)

CVE_2010_4478=$($dir/../jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion != "" and .CanonicalVersion <= "0005000600000000") | .IP' $tmpName | wc -l)
CVE_2008_3234=$($dir/../jq 'select(.Vendor=="OpenSSH" and .SoftwareVersion == "4.0" and select(.Comments | contains("Debian") and contains("ubuntu") == false) | .IP' $tmpName | wc -l)
CVE_2008_1657=$($dir/../jq 'select(.Vendor=="OpenSSH" and .CanonicalVersion <= "0004000900000000" and .CanonicalVersion >= "0004000400000000" and
 select(.Comments | endswith("5ubuntu0.6") == false and endswith("8ubuntu1.5") == false and endswith("7ubuntu3.5") == false)) | .IP' $tmpName | wc -l)


printf "%s\n" "CVE-2016-6515: $CVE_2016_6515"
printf "%s\n" "CVE-2015-8325: $CVE_2015_8325"
printf "%s\n" "CVE-2015-6565: $CVE_2015_6565"
printf "%s\n" "CVE-2015-6564: $CVE_2015_6564"
printf "%s\n" "CVE-2015-5600: $CVE_2015_5600"

printf "%s\n" "CVE-2014-1692: $CVE_2014_1692"
printf "%s\n" "CVE-2013-4548: $CVE_2013_4548"
printf "%s\n" "CVE-2010-4478: $CVE_2010_4478"
printf "%s\n" "CVE-2008-3234: $CVE_2008_3234"
printf "%s\n" "CVE-2008-1657: $CVE_2008_1657"


printf "Total: `$dir/../jq 'select(.Vendor=="OpenSSH") | .IP'  $tmpName | wc -l` \n"
printf '%s\n' '----------------------------------------------'
rm $tmpName