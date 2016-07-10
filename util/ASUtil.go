package util

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var maxMindASDB []asLiteEntry

const AS_FOLDER = "asdb"
const MAXMIND_AS_DB_ZIP_FILE_NAME = "GeoIPASNum2.zip"
const MAXMIND_AS_DB_FILE_NAME = "GeoIPASNum2.csv"

const MAXMIND_ASIP_URL = "http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNum2.zip"
const MAXMIND_AS_DB_ENTRY_REGEX = `(AS\d*)(?:\s(.*))?`

var maxmindASDBRegex = regexp.MustCompile(MAXMIND_AS_DB_ENTRY_REGEX)

type asLiteEntry struct {
	startIP int
	endIp   int
	asId    string
	asOwner string
}

func ASUtilInitialise() {
	if !CheckPathExist(AS_FOLDER) {
		err := os.MkdirAll(AS_FOLDER, FILE_ACCESS_PERMISSION)
		Check(err)
	}
	maxmindAsDBFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_FILE_NAME)
	maxmindAsDBZipFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_ZIP_FILE_NAME)
	if !CheckPathExist(maxmindAsDBFile) {
		downloadMaxMindASLite()
		Unzip(maxmindAsDBZipFile, AS_FOLDER)
		os.Remove(maxmindAsDBZipFile)
	} else if firstTuesdayOfMonth() {
		os.Remove(maxmindAsDBFile)
		downloadMaxMindASLite()
		Unzip(maxmindAsDBZipFile, AS_FOLDER)
		os.Remove(maxmindAsDBZipFile)
	}
	readInMaxMindASDBCSV()
}

func FindAS(ip string) (string, string) {
	if len(maxMindASDB) == 0 {
		panic(errors.New("ASDB(s) have not been initialised! Initialise first."))
	}
	if net.ParseIP(ip).To4() == nil {
		panic(errors.New(fmt.Sprintf("%v is not an IPv4 address\n", ip)))
	}

	ipToCheck := calculateMaxMindIpValue(ip)
	asId := "Not found"
	asOwner := "Not found"
	for _, v := range maxMindASDB {
		if ipToCheck >= v.startIP && ipToCheck <= v.endIp {
			asId = v.asId
			asOwner = v.asOwner
			break
		}
	}
	return asId, asOwner
}

func readInMaxMindASDBCSV() {
	maxmindAsDBFile := filepath.Join(AS_FOLDER, MAXMIND_AS_DB_FILE_NAME)

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
		startIP, errStart := strconv.Atoi(record[0])
		endIP, errEnd := strconv.Atoi(record[1])

		if errStart != nil {
			log.Printf("%v is not an IPv4 address\n", startIP)
			panic(errors.New("Wrong MaxMind GeoDB CSV file format!"))
		} else if errEnd != nil {
			log.Printf("%v is not an IPv4 address\n", endIP)
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

			entry := asLiteEntry{startIP: startIP, endIp: endIP, asId: asId, asOwner: asOwner}
			maxMindASDB = append(maxMindASDB, entry)
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

func calculateMaxMindIpValue(ip string) int {
	ipDigitsString := strings.Split(ip, ".")
	var ipDigits = []int{}

	for _, i := range ipDigitsString {
		j, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		ipDigits = append(ipDigits, j)
	}
	return (16777216 * ipDigits[0]) + (65536 * ipDigits[1]) + (256 * ipDigits[2]) + ipDigits[3]
}
