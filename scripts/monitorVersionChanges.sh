#!/bin/bash
nginxVersionUbuntu="1.10.0"
nginxVersionDebian="1.6.2"

apacheVersionUbuntu="2.4.18"
apacheVersionDebian="2.4.10"

scanRunning=`ps -e | grep zversion | wc -l`

if [ "$scanRunning" -le 1 ]
then
    #backup
    mv /etc/apt/sources.list /etc/apt/sources.list.3421temp.bak

    #create sources.list for debian jessie, just in case
    echo "deb http://httpredir.debian.org/debian jessie main" > /etc/apt/sources.list

    apt-get -qq update

    nginxLatestjessie=`apt-cache show nginx-common | grep Version | head -n 1 | egrep -o '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1`
    apacheLatestjessie=`apt-cache show apache2 | grep Version | head -n 1 | egrep -o '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1`

    #create sources.list for Ubuntu Xenial, just in case
    echo "deb http://de.archive.ubuntu.com/ubuntu xenial main restricted" > /etc/apt/sources.list
    echo "deb http://security.ubuntu.com/ubuntu xenial-security main restricted" >> /etc/apt/sources.list


    apt-get -qq update
    nginxLatestXenial=`apt-cache show nginx-common | grep Version | head -n 1 | egrep -o '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1`
    apacheLatestXenial=`apt-cache show apache2 | grep Version | head -n 1 | egrep -o '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1`

    #move backup back
    mv /etc/apt/sources.list.3421temp.bak /etc/apt/sources.list

    #update cache again correctly
    apt-get -qq update

    if [ "$nginxVersionUbuntu" !=  "$nginxLatestXenial" ] || [ "$nginxVersionDebian" !=  "$nginxLatestjessie" ] || [ "$apacheVersionUbuntu" !=  "$apacheLatestXenial" ] || [ "$apacheVersionDebian" !=  "$apacheLatestjessie" ]
    then
        ./zversion -hs -bf zmap_blacklist&
    fi
fi