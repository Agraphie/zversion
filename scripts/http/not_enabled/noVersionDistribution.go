package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/agraphie/zversion/http"
	"github.com/agraphie/zversion/ssh"
	"log"
	"os"
	"sort"
	"strconv"
)

var (
	file         = flag.String("file", "", "The file to analyse")
	serverVendor = flag.String("server-vendor", "", "The server vendor to look for")
	cmsVendor    = flag.String("cms-vendor", "", "The CMS vendor to look for")
	sshSoftware  = flag.String("ssh-software", "", "The SSH software to look for")
)

func init() {
	flag.StringVar(file, "f", "", "The file to analyse")
	flag.StringVar(serverVendor, "sv", "", "The server vendor to look for")
	flag.StringVar(cmsVendor, "cv", "", "The CMS vendor to look for")

	flag.Parse()

}

func main() {
	fmt.Println(*sshSoftware + *serverVendor + *cmsVendor + " no version distribution in " + *file)

	asnDistribution := map[string]int{}
	geoLocationDistribution := map[string]int{}
	versionUknown := 0

	buf := make([]byte, 0, 64*1024)

	file1, err1 := os.Open(*file)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer file1.Close()

	scanner1 := bufio.NewScanner(file1)
	scanner1.Buffer(buf, 1024*1024)
	var entry http.ZversionEntry
	var sshEntry ssh.SSHEntry

	for scanner1.Scan() {
		line := scanner1.Text()
		json.Unmarshal([]byte(line), &entry)

		if *cmsVendor != "" {
			if len(entry.CMS) > 0 {
				for _, v := range entry.CMS {
					if v.CanonicalVersion == "" && v.Vendor == *cmsVendor {
						geoLocationDistribution[entry.GeoData.RegisteredCountryCode]++
						asnDistribution[entry.ASId+"("+entry.ASOwner+")"]++
						versionUknown++
					}
				}
			}
		} else if *serverVendor != "" {
			if len(entry.Agents) > 0 {
				for _, v := range entry.Agents {
					if v.CanonicalVersion == "" && v.Vendor == *serverVendor {
						geoLocationDistribution[entry.GeoData.RegisteredCountryCode]++
						asnDistribution[entry.ASId+"("+entry.ASOwner+")"]++
						versionUknown++
					}
				}
			}
		} else if *sshSoftware != "" {
			json.Unmarshal([]byte(line), &sshEntry)
			if sshEntry.SoftwareVersion == "" && sshEntry.Vendor == *sshSoftware {
				geoLocationDistribution[entry.GeoData.RegisteredCountryCode]++
				asnDistribution[entry.ASId+"("+entry.ASOwner+")"]++
				versionUknown++
			}
		}
	}

	if err1 := scanner1.Err(); err1 != nil {
		log.Fatal(err1)
	}
	asnDistributionSort := rankByWordCount(asnDistribution)
	geoLocationDistributionSort := rankByWordCount(geoLocationDistribution)

	fmt.Printf("-----------Version unknown ASN distribution (%v in total)--------------\n", versionUknown)
	for _, k := range asnDistributionSort {
		fmt.Println(strconv.Itoa(k.Value) + " " + k.Key)
	}
	fmt.Println("--------------------------------------")
	fmt.Println("")

	fmt.Println("-----------Registered geo location distribution-----------")
	for _, k := range geoLocationDistributionSort {
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
