package util

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

var maxMindGeoDB []geoLiteEntry

const GEODB_FOLDER = "geodb"
const MAXMIND_GEO_DB_ZIP_FILE_NAME = "GeoIPCountryCSV.zip"
const MAXMIND_GEO_DB_FILE_NAME = "GeoIPCountryWhois.csv"

const MAXMIND_GEOIP_URL = "http://geolite.maxmind.com/download/geoip/database/GeoIPCountryCSV.zip"

type geoLiteEntry struct {
	startIP     net.IP
	endIp       net.IP
	countryCode string
}

func GeoUtilInitialise() {
	if !CheckPathExist(GEODB_FOLDER) {
		err := os.MkdirAll(GEODB_FOLDER, FILE_ACCESS_PERMISSION)
		Check(err)
	}
	if !CheckPathExist(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_FILE_NAME) {
		downloadMaxMindGeoLite()
		Unzip(GEODB_FOLDER+"/"+MAXMIND_GEO_DB_ZIP_FILE_NAME, GEODB_FOLDER)
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_ZIP_FILE_NAME)
	} else if firstTuesdayOfMonth() {
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_FILE_NAME)
		downloadMaxMindGeoLite()
		Unzip(GEODB_FOLDER+"/"+MAXMIND_GEO_DB_ZIP_FILE_NAME, GEODB_FOLDER)
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_ZIP_FILE_NAME)
	}
	readInMaxMindGeoDBCSV(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_FILE_NAME)
}

func FindCountry(ip string) string {
	if len(maxMindGeoDB) == 0 {
		panic(errors.New("GeoDB(s) have not been initialised! Initialise first."))
	}
	countryCode := "Not found"
	ipToCheck := net.ParseIP(ip)
	if ipToCheck.To4() == nil {
		panic(errors.New(fmt.Sprintf("%v is not an IPv4 address\n", ipToCheck)))
	}
	for _, v := range maxMindGeoDB {
		if bytes.Compare(ipToCheck, v.startIP) >= 0 && bytes.Compare(ipToCheck, v.endIp) <= 0 {
			countryCode = v.countryCode
			break
		}
	}
	return countryCode
}

func readInMaxMindGeoDBCSV(filePath string) {
	f, err := os.Open(filePath)
	defer f.Close()
	Check(err)
	reader := csv.NewReader(bufio.NewReader(f))
	for {
		record, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		startIP := net.ParseIP(record[0])
		endIP := net.ParseIP(record[1])
		if startIP.To4() == nil {
			log.Printf("%v is not an IPv4 address\n", startIP)
			panic(errors.New("Wrong MaxMind GeoDB CSV file format!"))
		} else if endIP.To4() == nil {
			log.Printf("%v is not an IPv4 address\n", endIP)
			panic(errors.New("Wrong MaxMind GeoDB CSV file format!"))
		}

		entry := geoLiteEntry{startIP: startIP, endIp: endIP, countryCode: record[4]}
		maxMindGeoDB = append(maxMindGeoDB, entry)
	}
}

func downloadMaxMindGeoLite() {
	out, err := os.Create(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_ZIP_FILE_NAME)
	defer out.Close()
	Check(err)

	resp, err1 := http.Get(MAXMIND_GEOIP_URL)
	defer resp.Body.Close()
	Check(err1)

	_, err2 := io.Copy(out, resp.Body)
	Check(err2)
}
