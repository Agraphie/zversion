package util

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/thekvs/go-net-radix"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var maxMindASDB []asLiteEntry

const AS_FOLDER = "asdb"
const MAXMIND_AS_DB_ZIP_FILE_NAME = "GeoIPASNum2.zip"
const MAXMIND_AS_DB_FILE_NAME = "GeoIPASNum2.csv"

const MAXMIND_ASIP_URL = "http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNum2.zip"
const MAXMIND_AS_DB_ENTRY_REGEX = `(AS\d*)(?:\s(.*))?`

var maxmindASDBRegex = regexp.MustCompile(MAXMIND_AS_DB_ENTRY_REGEX)
var asnRadixTree *netradix.NetRadixTree

type asLiteEntry struct {
	asId    string
	asOwner string
}

func ASUtilInitialise() {
	if !CheckPathExist(AS_FOLDER) {
		err := os.MkdirAll(AS_FOLDER, FILE_ACCESS_PERMISSION)
		Check(err)
	}
	defer TimeTrack(time.Now(), "Initialising ASN DB")

	maxmindAsDBFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_FILE_NAME)
	maxmindAsDBZipFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_ZIP_FILE_NAME)
	if !CheckPathExist(maxmindAsDBFile) {
		downloadMaxMindASLite()
		Unzip(maxmindAsDBZipFile, AS_FOLDER)
		os.Remove(maxmindAsDBZipFile)
	} else {
		info, err := os.Stat(maxmindAsDBFile)
		Check(err)
		fileOldMonth := info.ModTime().Month() < time.Now().Month()
		fileOldDay := info.ModTime().Day() >= SecondTuesday()

		if fileOldMonth || fileOldDay {
			os.Remove(maxmindAsDBFile)
			downloadMaxMindASLite()
			Unzip(maxmindAsDBZipFile, AS_FOLDER)
			os.Remove(maxmindAsDBZipFile)
		}
	}
	readInMaxMindASDBCSV()
}

func FindAS(ip string) (string, string) {
	if asnRadixTree == nil {
		panic(errors.New("ASDB(s) have not been initialised! Initialise first."))
	}
	asId := "Not found"
	asOwner := "Not found"

	ipToCheck := net.ParseIP(ip)
	if ipToCheck.To4() == nil {
		log.Print(errors.New(fmt.Sprintf("%v is not an IPv4 address\n", ipToCheck)))
	} else {
		found, udata, err := asnRadixTree.SearchBest(ip)
		Check(err)

		if found {
			split := strings.Split(udata, ",")
			asId = split[0]
			asOwner = split[1]
		}
	}
	return asId, asOwner
}

func readInMaxMindASDBCSV() {
	maxmindAsDBFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_FILE_NAME)
	var err1 error
	asnRadixTree, err1 = netradix.NewNetRadixTree()
	Check(err1)

	f, err := os.Open(maxmindAsDBFile)
	defer f.Close()
	Check(err)
	reader := csv.NewReader(bufio.NewReader(f))
	for {
		record, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		startIPInt, errStart := strconv.Atoi(record[0])
		endIPInt, errEnd := strconv.Atoi(record[1])

		if errStart != nil {
			log.Printf("%v is not an IPv4 address\n", startIPInt)
			panic(errors.New("Wrong MaxMind GeoDB CSV file format!"))
		} else if errEnd != nil {
			log.Printf("%v is not an IPv4 address\n", endIPInt)
			panic(errors.New("Wrong MaxMind GeoDB CSV file format!"))
		}

		asInfo := maxmindASDBRegex.FindStringSubmatch(record[2])
		if len(asInfo) == 0 {
			panic(errors.New("No AS given for entry: " + record[0] + "," + record[1]))
		} else {
			asId := asInfo[1]
			asOwner := "Not found"
			if len(asInfo) > 2 {
				asOwner = asInfo[2]
			}
			startIP := calculateIPFromMaxmindValue(startIPInt)
			endIP := calculateIPFromMaxmindValue(endIPInt)

			cidr := calculateCidr(startIP, endIP)

			asnRadixTree.Add(cidr.String(), asId+","+asOwner)

			//maxMindASDB = append(maxMindASDB, entry)
		}
	}
}

func downloadMaxMindASLite() {
	maxmindAsDBZipFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_ZIP_FILE_NAME)

	out, err := os.Create(maxmindAsDBZipFile)
	defer out.Close()
	Check(err)

	resp, err1 := http.Get(MAXMIND_ASIP_URL)
	defer resp.Body.Close()
	Check(err1)

	_, err2 := io.Copy(out, resp.Body)
	Check(err2)
}

func calculateIPFromMaxmindValue(ipnum int) net.IP {
	o1 := (ipnum / 16777216) % 256
	o2 := (ipnum / 65536) % 256
	o3 := (ipnum / 256) % 256
	o4 := ipnum % 256

	ip := strconv.Itoa(o1) + "." + strconv.Itoa(o2) + "." + strconv.Itoa(o3) + "." + strconv.Itoa(o4)

	return net.ParseIP(ip)
}

func calculateCidr(ip1, ip2 net.IP) *net.IPNet {
	maxLen := 32

	for l := maxLen; l >= 0; l-- {
		mask := net.CIDRMask(l, maxLen)
		na := ip1.Mask(mask)
		n := net.IPNet{IP: na, Mask: mask}

		if n.Contains(ip2) {
			muh := strconv.Itoa(l)
			_, result1, err := net.ParseCIDR(na.String() + "/" + muh)
			Check(err)
			return result1
		}
	}
	return nil
}
