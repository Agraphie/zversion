package util

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

var maxMindGeoDB []geoLiteEntry

const GEODB_FOLDER = "geodb"
const MAXMIND_GEO_DB_ZIP_FILE_NAME = "GeoLite2-Country-CSV.zip"
const MAXMIND_GEO_DB_FILE_NAME = "GeoLite2-Country-Blocks-IPv4.csv"
const MAXMIND_GEO_DB_COUNTRY_CODES_FILE_NAME = "GeoLite2-Country-Locations-en.csv"

const MAXMIND_GEOIP_URL = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country-CSV.zip"

var countries map[string]countryEntry

type countryEntry struct {
	countryCode   string
	continentCode string
}

type geoLiteEntry struct {
	network                  *net.IPNet
	registeredCountryCode    string
	geolocationCountryCode   string
	registeredContinentCode  string
	geolocationContinentCode string
}

type GeoData struct {
	RegisteredCountryCode    string
	GeolocationCountryCode   string
	RegisteredContinentCode  string
	GeolocationContinentCode string
}

func GeoUtilInitialise() {
	if !CheckPathExist(GEODB_FOLDER) {
		err := os.MkdirAll(GEODB_FOLDER, FILE_ACCESS_PERMISSION)
		Check(err)
	}
	if !CheckPathExist(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_FILE_NAME) {
		downloadMaxMindGeoLite()
		err := Unzip(GEODB_FOLDER+"/"+MAXMIND_GEO_DB_ZIP_FILE_NAME, GEODB_FOLDER)
		Check(err)
		maxMindCleanUp()
	} else if firstTuesdayOfMonth() {
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_FILE_NAME)
		downloadMaxMindGeoLite()
		err := Unzip(GEODB_FOLDER+"/"+MAXMIND_GEO_DB_ZIP_FILE_NAME, GEODB_FOLDER)
		Check(err)
		maxMindCleanUp()
	}
	readInMaxMindGeoDBCSV()
}

func maxMindCleanUp() {
	os.Remove(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_ZIP_FILE_NAME)
	d, err := os.Open(GEODB_FOLDER)
	Check(err)

	defer d.Close()
	names, err := d.Readdirnames(-1)
	Check(err)

	for _, name := range names {
		if name != MAXMIND_GEO_DB_COUNTRY_CODES_FILE_NAME && name != MAXMIND_GEO_DB_FILE_NAME {
			err := os.Remove(filepath.Join(GEODB_FOLDER, name))
			Check(err)
		}
	}
}
func FindGeoData(ip string) GeoData {
	if len(maxMindGeoDB) == 0 {
		panic(errors.New("GeoDB(s) have not been initialised! Initialise first."))
	}
	registeredCountryCode := "Not found"
	geolocationCountryCode := "Not found"
	registeredContinentCode := "Not found"
	geolocationContinentCode := "Not found"

	ipToCheck := net.ParseIP(ip)
	if ipToCheck.To4() == nil {
		panic(errors.New(fmt.Sprintf("%v is not an IPv4 address\n", ipToCheck)))
	}
	for _, v := range maxMindGeoDB {
		if v.network.Contains(ipToCheck) {
			registeredCountryCode = v.registeredCountryCode
			geolocationCountryCode = v.geolocationCountryCode
			registeredContinentCode = v.registeredContinentCode
			geolocationContinentCode = v.geolocationContinentCode
			break
		}
	}

	return GeoData{registeredCountryCode, geolocationCountryCode,
		registeredContinentCode, geolocationContinentCode}
}

func readInMaxMindGeoDBCSV() {
	readInCountryCodes()
	dbCsv, err := os.Open(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_FILE_NAME)
	defer dbCsv.Close()
	Check(err)

	reader := csv.NewReader(bufio.NewReader(dbCsv))
	//Skip header line
	reader.Read()
	for {
		record, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		_, network, errIp := net.ParseCIDR(record[0])
		Check(errIp)
		registeredCountryId := record[1]
		representedCountryId := record[2]

		entry := geoLiteEntry{network: network, registeredCountryCode: countries[registeredCountryId].countryCode, geolocationCountryCode: countries[representedCountryId].countryCode,
			registeredContinentCode: countries[registeredCountryId].continentCode, geolocationContinentCode: countries[representedCountryId].continentCode}
		maxMindGeoDB = append(maxMindGeoDB, entry)
	}
}

func readInCountryCodes() {
	dbCodesCsv, err1 := os.Open(GEODB_FOLDER + "/" + MAXMIND_GEO_DB_COUNTRY_CODES_FILE_NAME)
	defer dbCodesCsv.Close()
	Check(err1)
	reader := csv.NewReader(bufio.NewReader(dbCodesCsv))
	countries = make(map[string]countryEntry)

	for {
		record, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		country := countryEntry{countryCode: record[4], continentCode: record[2]}
		countries[record[0]] = country
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
