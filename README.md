# ZVersion

ZVersion is a tool for analysing [ZGrab](https://github.com/zmap/zgrab) output, programmed as part of a Master thesis at the [COMSYS](www.comsys.rwth-aachen.de) chair of [RWTH Aachen](http://www.rwth-aachen.de). So far, ZVersion handles HTTP and SSH data sets. Data sets for analysing can be found in the in the project [Sonar repositories](https://scans.io/study/sonar.http) and [Censys repositories](https://censys.io/data).

ZVersion will extract the software and versions from the input data set. It will also enrich every IP them with Geo- and ASN-data automatically.

### Analysis Execution
Depending on the used computer and input file size, the analysis can take several hours. Please use a small data set for testing, before blocking your CPU.
After obtaining a data set from the above sources, a ZVersion analysis can be simply executed for HTTP with
```bash
./zversion -ha -ai http-input-file-censys.json.lz4
```
or
```bash
./zversion -ha -ai http-input-file-rapid7.gz
```

and for SSH with

```bash
./zversion -sa -ai ssh-input-file-censys.json.lz4
```

All scripts which are in the folder scripts/ssh i.e. scripts/http get called with the first and only parameter being the ZVersion output file which is under analysis. If you wish to have new analyses executed automatically, just drop a script or program in the respective folder. It will then get executed for every analysis.

For all further usage of ZVersion please refrain to 
```bash
./zversion --help
```

### Scan Execution
SSH and HTTP scans can be performed, but only with the help of [ZMap](https://zmap.io/) and ZGrab. Thus, ZGrab and ZMap have to be installed on the scanning machine. Please make sure that the ZMap and ZGrab work correctly before continuing. In order to launch a restricted scan i.e. only specific IPs should be scanned, please provide an input file to ZVersion, where one line is one IP. VHost scans can be launched by specifing the input file as follows
```
137.226.107.63,www.rwth-aachen.de
```

A vhost HTTP scan can then be launched with
```
./zversio -hs -si inputFile
```
All results are put in scanResults/ssh i.e. scanResults/http. For further usage, please refrain to ```./zversion --help```

### Compilation
In order to compile ZVersion, Go needs to be set up correctly. Additionally, ZVersion makes use of these two projects, which need to be cloned
https://github.com/oschwald/maxminddb-golang

https://github.com/armon/go-radix

After obtaining these dependencies, ZVersion can be compiled with ```go build zversion.go```.
