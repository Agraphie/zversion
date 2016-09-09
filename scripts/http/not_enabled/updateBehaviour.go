package main

import (
	"bufio"
	"encoding/json"
	"errors"
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
	oldFile          = flag.String("old-file", "", "The old file to compare")
	newFile          = flag.String("new-file", "", "The new file to compare")
	serverVendor     = flag.String("server-vendor", "", "The server vendor to look for")
	cmsVendor        = flag.String("cms-vendor", "", "The CMS vendor to look for")
	sshVendor        = flag.String("ssh-vendor", "", "The SSH vendor to look for")
	canonicalVersion = flag.String("version", "", "The version in canonical form. Use this for SSH and HTTP")
	stringVersion    = flag.String("string-version", "", "The version as string. Use only for SSH! Not for HTTP!")
	stableIPs        = flag.String("stable-ips", "", "Path to the file with the stable IPs")
)

func init() {
	flag.StringVar(oldFile, "of", "", "The old file to compare")
	flag.StringVar(newFile, "nf", "", "The new file to compare")
	flag.StringVar(serverVendor, "sv", "", "The server vendor to look for")
	flag.StringVar(cmsVendor, "cv", "", "The CMS vendor to look for")
	flag.StringVar(sshVendor, "sshv", "", "The SSH vendor to look for")
	flag.StringVar(canonicalVersion, "v", "", "The version in canonical form")
	//	flag.StringVar(stringVersion, "stringv", "", "The version in canonical form")
	flag.StringVar(stableIPs, "si", "", "Path to the file with the stable IPs")
	flag.Parse()

}

func main() {
	fmt.Println(*serverVendor + *cmsVendor + *sshVendor + " upgraded from versions in " + *oldFile + " to version " + *stringVersion + *canonicalVersion + " in " + *newFile)
	ips := make(map[string]string)
	if *stableIPs != "" {
		ipsWhole, err := readLines(*stableIPs)
		if err != nil {
			panic(err)
		}
		for _, value := range ipsWhole {
			ips[value] = value
		}
	}

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

	var entry http.ZversionEntry
	var sshEntry ssh.SSHEntry
	for scanner.Scan() {
		line := scanner.Bytes()

		if *cmsVendor != "" {
			json.Unmarshal(line, &entry)
			if len(ips) > 0 {
				if _, ok := ips[entry.IP]; !ok {
					continue
				}
			}
			if len(entry.CMS) > 0 {
				for _, v := range entry.CMS {
					if v.Vendor == *cmsVendor && v.CanonicalVersion == *canonicalVersion {
						updateCount++
						entries[entry.IP] = v.Version
					}
				}
			}
		} else if *serverVendor != "" {
			json.Unmarshal(line, &entry)
			if len(ips) > 0 {
				if _, ok := ips[entry.IP]; !ok {
					continue
				}
			}
			if len(entry.Agents) > 0 {
				for _, v := range entry.Agents {
					if v.Vendor == *serverVendor && v.CanonicalVersion == *canonicalVersion {
						updateCount++
						entries[entry.IP] = v.Version
					}
				}
			}
		} else if *sshVendor != "" {
			json.Unmarshal(line, &sshEntry)
			if len(ips) > 0 {
				if _, ok := ips[sshEntry.IP]; !ok {
					continue
				}
			}
			if sshEntry.Vendor == *sshVendor {
				if *stringVersion != "" {
					if sshEntry.SoftwareVersion == *stringVersion {
						updateCount++
						entries[sshEntry.IP] = sshEntry.SoftwareVersion
					}
				} else {
					if sshEntry.CanonicalVersion == *canonicalVersion {
						updateCount++
						entries[sshEntry.IP] = sshEntry.SoftwareVersion
					}
				}
			}
		} else {
			panic(errors.New("I don't understand..."))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file1, err1 := os.Open(*oldFile)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer file1.Close()

	scanner1 := bufio.NewScanner(file1)
	scanner1.Buffer(buf, 1024*1024)

	ipsFound := 0
	for scanner1.Scan() {
		line := scanner1.Bytes()
		json.Unmarshal(line, &entry)

		if _, ok := entries[entry.IP]; ok {
			ipsFound += 1
			if *cmsVendor != "" {
				if len(entry.CMS) > 0 {
					for _, v := range entry.CMS {
						if v.Vendor == *cmsVendor && v.CanonicalVersion != *canonicalVersion {
							sum[v.Version]++
							asn[entry.ASId+"("+entry.ASOwner+")"]++
						}
					}
				}
			} else if *serverVendor != "" {
				if len(entry.Agents) > 0 {
					for _, v := range entry.Agents {
						if v.Vendor == *serverVendor && v.CanonicalVersion != *canonicalVersion {
							sum[v.Version]++
							asn[entry.ASId+"("+entry.ASOwner+")"]++
						}
					}
				}
			} else if *sshVendor != "" {
				json.Unmarshal(line, &sshEntry)
				if sshEntry.Vendor == *sshVendor {
					if *stringVersion != "" {
						if sshEntry.SoftwareVersion != *stringVersion {
							sum[sshEntry.SoftwareVersion]++
							asn[entry.ASId+"("+entry.ASOwner+")"]++
						}
					} else {
						if sshEntry.CanonicalVersion != *canonicalVersion {
							sum[sshEntry.SoftwareVersion]++
							asn[entry.ASId+"("+entry.ASOwner+")"]++
						}
					}
				}
			} else {
				panic(errors.New("I don't understand..."))
			}
		}
	}

	notFoundIPsLength := len(entries) - ipsFound

	sum["IPs not found"] = notFoundIPsLength
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

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
