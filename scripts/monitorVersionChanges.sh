#!/bin/bash
nginxVersionUbuntu="1.10.1"
nginxVersionDebian="1.6.2"

apacheVersionUbuntu="2.4.18"
apacheVersionDebian="2.4.10"

openSSHUbuntu="7.2p2"
openSSHDebian="6.7p1"
scanRunning=$(pgrep -c zversion)

if [ "$scanRunning" -le 1 ]
then
    nginxLatestjessie=$(curl -s https://packages.debian.org/jessie/nginx-full | grep '<li><a href="http://ftp-master.metadata.debian.org/changelogs//main/n/nginx/' |  sed -n  's/.*\([[:digit:]]\+\.[[:digit:]]\+\.[[:digit:]]\+\).*/\1/p' | head -n 1)
    apacheLatestjessie=$(curl -s https://packages.debian.org/jessie/apache2 | grep '<li><a href="http://ftp-master.metadata.debian.org/changelogs//main/a/apache2/' |  sed -n  's/.*\([[:digit:]]\+\.[[:digit:]]\+\.[[:digit:]]\+\).*/\1/p'| head -n 1)
    openSSHServerLatestjessie=$(curl -s https://packages.debian.org/jessie/openssh-server | grep '<li><a href="http://ftp-master.metadata.debian.org/changelogs//main/o/openssh/' |  sed -n  's/.*\([[:digit:]]\+\.[[:digit:]]\+p[[:digit:]]\).*/\1/p' | head -n 1)

    nginxLatestXenial=$(curl -s https://launchpad.net/ubuntu/xenial/+source/nginx | grep '<h2>Download files from current release' | sed -n  's/.*\([[:digit:]]\+\.[[:digit:]]\+\.[[:digit:]]\+\).*ubuntu.*/\1/p' | head -n 1)
    apacheLatestXenial=$(curl -s https://launchpad.net/ubuntu/xenial/+source/apache2 | grep '<h2>Download files from current release' | sed -n  's/.*\([[:digit:]]\+\.[[:digit:]]\+\.[[:digit:]]\+\).*ubuntu.*/\1/p' | head -n 1)
    openSSHServerUbuntu=$(curl -s https://launchpad.net/ubuntu/xenial/+source/openssh | grep '<h2>Download files from current release' | sed -n  's/.*\([[:digit:]]\+\.[[:digit:]]\+p[[:digit:]]\).*/\1/p' | head -n 1)

    if [ "$nginxVersionUbuntu" !=  "$nginxLatestXenial" ]
    then
        echo "Expected version $nginxVersionUbuntu but got $nginxLatestXenial for Ubuntu"
    fi

    if [ "$nginxVersionDebian" !=  "$nginxLatestjessie" ]
    then
            echo "Expected version $nginxVersionDebian but got $nginxLatestjessie for Debian"
    fi

    if [ "$apacheVersionUbuntu" !=  "$apacheLatestXenial" ]
    then
            echo "Expected version $apacheVersionUbuntu but got $apacheLatestXenial for Ubuntu"
    fi

    if  [ "$openSSHUbuntu" !=  "$openSSHServerUbuntu" ]
    then
            echo "Expected version $openSSHUbuntu but got $openSSHServerUbuntu for Ubuntu"
    fi

    if  [ "$apacheVersionDebian" !=  "$apacheLatestjessie" ]
    then
            echo "Expected version $apacheVersionDebian but got $apacheLatestjessie for Debian"
    fi

    if  [ "$openSSHDebian" !=  "$openSSHServerLatestjessie" ]
    then
            echo "Expected version $openSSHDebian but got $openSSHServerLatestjessie for Debian"
    fi

fi