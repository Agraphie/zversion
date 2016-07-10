#!/bin/bash
#Output file name: OpenSSH_CVE_2016-0777_and_2016-0778
printf $1
printf "\n"
printf "5.4 >= OpenSSH Version <= 7.1: `jq 'select(.Vendor =="OpenSSH" and .CanonicalVersion <= "0007000100000000" and .CanonicalVersion >= "0005000400000000" and .SoftwareVersion != "" and .SoftwareVersion != "7.1p1") | .SoftwareVersion' $1 | wc -l`"
printf "\n"
printf "Total OpenSSH: `jq 'select(.Vendor =="OpenSSH") | .Vendor' $1 | wc -l`"
printf "\n"
