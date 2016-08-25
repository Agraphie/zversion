#!/bin/bash

pattern=$1
mode=$2

if [ "$mode" == "ssh" ];
then
    rm sshStableIps
    touch sshStableIps

    for i in cleanedResults/ssh/"$pattern"_*/ssh_version.json; do
        while read line; do
            ip=$(echo $line | jq '.IP')
            if grep -q "$ip" sshStableIps; then
                continue
            fi
            vendor=$(echo $line | jq '.Vendor')
            count=$(grep "$ip" cleanedResults/ssh/"$pattern"_*/ssh_version.json | grep "$vendor" | wc -l)

            if (( $count >= 4 )); then
                echo $ip >> sshStableIps
            fi
       done <  $i
    done
fi
if [ "$mode" == "http" ];
then
    rm httpStableIps
    touch httpStableIps
    for i in cleanedResults/http/"$pattern"_*/http_version.json; do
            while read line; do
                ip=$(echo $line | jq '.IP')
                if grep -q "$ip" httpStableIps; then
                    continue
                else
                    vendor=$(echo $line | jq '.Agents[].Vendor')
                    count=$(grep "$ip" cleanedResults/http/"$pattern"_*/http_version.json | grep "$vendor" | wc -l)
                    if (( $count >= 4 )); then
                        echo $ip >> httpStableIps
                    fi
                fi

            done < $i
    done
fi
