package main

import (
	"bufio"
	"encoding/json"
	"github.com/agraphie/zversion/http1"
	"log"
	"os"
)

func main() {
	entries := map[string]string{}
	sum := map[string]int{}

	filepathNew := os.Args[1]
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
				if v.Vendor == "WordPress" && v.CanonicalVersion >= "0004000500030000" {
					entries[entry.IP] = v.Version
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	filepathOld := os.Args[2]
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
					if v.Vendor == "WordPress" {
						sum[v.Version]++
					}
				}
			}
		}
	}

	if err1 := scanner1.Err(); err1 != nil {
		log.Fatal(err1)
	}
	log.Println("Upgraded from versions in " + filepathOld + " to version 4.5.3 in " + filepathNew)
	log.Println(sum)
}
