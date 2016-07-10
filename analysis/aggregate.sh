#!/bin/bash
echo "Server;Count" >> aggregated.csv;
serverNames=( "httpserver" "kangle" "openresty" "smartcds" "tinyproxy" "Virata-EmWeb" "WebSphereApplicationServer" "Zeus" "Zope" "fec\/1.0(FunkwerkBOSS)" "gunicorn" "SiemensGigaset-Server" "Simple,SecureWebServer1\.1" "SonicWALLSSL-VPNWebServer" "Unknown\/0.0UPnP\/1.0Conexant-EmWeb\/R6_1_0" "Resin" "RomPager.*UPNP" "RouterWebConfigSystem" "RouterWebServer" "Radware-web-server" "Rapidsite\/Apa\/1.3.33(Unix)FrontPage\/5.0.2.2510mod_ssl\/2.8.22OpenSSL\/0.9.8d" "RAC_ONE_HTTP" "RGOSHTTP-Server" "Mbedthis-AppWeb" "Microsoft-HTTPAPI\/1\.0" "Microsoft-HTTPAPI\/2\.0" "PanasonicAVCServer" "Miniwebserver1\.0ZTEcorp2005\." "KM-MFP-http" "Akamai" "Kerio" "Linux,HTTP\/1\.1,D" "GlassFish" "HPHTTPServer" "Hikvision-Webs" "Indy" "ECAcc" "ECD" "ECS" "EmbeddedHTTPServer" "D-LinkWebServer" "Cougar" "CommuniGatePro" "Boa" "BarracudaHTTP" "AnyLink" "Not set" "Error" "Oracle-Application-Server" "Apache" "Nginx" "Microsoft.IIS" "ATS" "Abyss" "Allegro" "Tengine" "squid" "httpd" "Jetty" "webserver") 
total=0
for i in "${serverNames[@]}"
do
	b=`sed -n "/\${i}/Ip" $1 | awk -F ":" '{SUM += $2} END {print SUM}'`
        echo "$i;$b" >> aggregated.csv
	(( total+=b ))
done	
totalProcessed=`grep "Total Processed" $1 | awk -F ":" '{print $2}' | sed 's/,/ /g'`
echo $total of $totalProcessed
