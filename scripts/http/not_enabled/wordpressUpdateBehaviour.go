package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/http1"
	"log"
	"os"
	"strconv"
)

func main() {
	entries := map[string]string{}
	sum := map[string]int{}
	asn := map[string]int{}

	filepathNew := os.Args[1]
	filepathOld := os.Args[2]
	vendor := os.Args[3]
	canonicalVersion := os.Args[4]
	file, err := os.Open(filepathNew)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	buf := make([]byte, 0, 64*1024)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, 1024*1024)

	var entry http1.ZversionEntry
	for scanner.Scan() {
		line := scanner.Text()
		json.Unmarshal([]byte(line), &entry)

		if len(entry.CMS) > 0 {
			for _, v := range entry.CMS {
				if v.Vendor == vendor && v.CanonicalVersion >= canonicalVersion {
					entries[entry.IP] = v.Version
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file1, err1 := os.Open(filepathOld)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer file.Close()

	scanner1 := bufio.NewScanner(file1)
	scanner1.Buffer(buf, 1024*1024)

	for scanner1.Scan() {
		line := scanner1.Text()
		json.Unmarshal([]byte(line), &entry)

		if _, ok := entries[entry.IP]; ok {
			if len(entry.CMS) > 0 {
				for _, v := range entry.CMS {
					if v.Vendor == vendor {
						sum[v.Version]++
						asn[entry.ASId+"("+entry.ASOwner+")"]++
					}
				}
			}
		}
	}

	if err1 := scanner1.Err(); err1 != nil {
		log.Fatal(err1)
	}
	log.Println(vendor + " upgraded from versions in " + filepathOld + " to version " + canonicalVersion + " in " + filepathNew)
	log.Println("-----------Upgraded from--------------")
	for v, k := range sum {
		fmt.Println(strconv.Itoa(k) + " " + v)
	}
	log.Println("--------------------------------------")

	log.Println("-----------ASN distribution-----------")
	for v, k := range asn {
		fmt.Println(strconv.Itoa(k) + " " + v)
	}
	log.Println("--------------------------------------")
}
