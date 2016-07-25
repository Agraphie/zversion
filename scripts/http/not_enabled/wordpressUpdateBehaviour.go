package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/agraphie/zversion/http1"
	"log"
	"os"
	"sort"
	"strconv"
)

var (
	oldFile          = flag.String("old-file", "", "The old file to compare")
	newFile          = flag.String("new-file", "", "The new file to compare")
	serverVendor     = flag.String("server-vendor", "", "The server vendor to look for")
	cmsVendor        = flag.String("cms-vendor", "", "The CMS vendor to look for")
	canonicalVersion = flag.String("version", "", "The version in canonical form")
)

func init() {
	flag.StringVar(oldFile, "of", "", "The old file to compare")
	flag.StringVar(newFile, "nf", "", "The new file to compare")
	flag.StringVar(serverVendor, "sv", "", "The server vendor to look for")
	flag.StringVar(cmsVendor, "cv", "", "The CMS vendor to look for")
	flag.StringVar(canonicalVersion, "v", "", "The version in canonical form")

	flag.Parse()

}

func main() {
	fmt.Println(*serverVendor + *cmsVendor + " upgraded from versions in " + *oldFile + " to version " + *canonicalVersion + " in " + *newFile)

	entries := map[string]string{}
	sum := map[string]int{}
	asn := map[string]int{}
	updateCount := 0

	file, err := os.Open(*newFile)
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

		if *cmsVendor != "" {
			if len(entry.CMS) > 0 {
				for _, v := range entry.CMS {

					if v.Vendor == *cmsVendor && v.CanonicalVersion == *canonicalVersion {
						updateCount++
						entries[entry.IP] = v.Version
					}
				}
			}
		} else {
			if len(entry.Agents) > 0 {
				for _, v := range entry.Agents {
					if v.Vendor == *serverVendor && v.CanonicalVersion == *canonicalVersion {
						updateCount++
						entries[entry.IP] = v.Version
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file1, err1 := os.Open(*oldFile)
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
			if *cmsVendor != "" {
				if len(entry.CMS) > 0 {
					for _, v := range entry.CMS {
						if v.Vendor == *cmsVendor {
							sum[v.Version]++
							asn[entry.ASId+"("+entry.ASOwner+")"]++
						}
					}
				}
			} else {
				if len(entry.Agents) > 0 {
					for _, v := range entry.Agents {
						if v.Vendor == *serverVendor {
							sum[v.Version]++
							asn[entry.ASId+"("+entry.ASOwner+")"]++
						}
					}
				}
			}
		}
	}

	if err1 := scanner1.Err(); err1 != nil {
		log.Fatal(err1)
	}
	versionSort := rankByWordCount(sum)
	asnSort := rankByWordCount(asn)

	fmt.Printf("-----------Upgraded from (%v in total)--------------\n", updateCount)
	for _, k := range versionSort {
		fmt.Println(strconv.Itoa(k.Value) + " " + k.Key)
	}
	fmt.Println("--------------------------------------")
	fmt.Println("")

	fmt.Println("-----------ASN distribution-----------")
	for _, k := range asnSort {
		fmt.Println(strconv.Itoa(k.Value) + " " + k.Key)
	}
	fmt.Println("--------------------------------------")
}

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
