package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//var maxMindGeoDB []geoLiteEntry

type Asc []geoLiteEntry

func (s Asc) Len() int {
	return len(s)
}
func (s Asc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Asc) Less(i, j int) bool {
	return bytes.Compare(s[i].network.IP, s[j].network.IP) == -1
}

const GEODB_FOLDER = "geodb"
const MAXMIND_GEO_DB_ZIP_FILE_NAME = "GeoLite2-Country.mmdb.gz"
const MAXMIND_GEO_DB_FILE_NAME = "GeoLite2-Country.mmdb"

//const MAXMIND_GEO_DB_COUNTRY_CODES_FILE_NAME = "GeoLite2-Country-Locations-en.csv"

const MAXMIND_GEOIP_URL = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz"

var countries map[string]countryEntry
var maxMindGeoDB *maxminddb.Reader

type countryEntry struct {
	countryCode   string
	continentCode string
}

type geoLiteEntry struct {
	network                  *net.IPNet
	firstIP                  string
	lastIP                   string
	registeredCountryCode    string
	geolocationCountryCode   string
	registeredContinentCode  string
	geolocationContinentCode string
}

func (c geoLiteEntry) String() string {
	return fmt.Sprintf("Network: %s, Continent: %s,Last IP: %s, First IP: %s", c.network, c.registeredContinentCode, c.lastIP, c.firstIP)
}

type GeoData struct {
	RegisteredCountryCode  string `json:"RegCountry" maxminddb:"iso_code"`
	GeolocationCountryCode string `json:"GeoCountry"`
	Continent              string `json:"Continent"`
}

type geoData struct {
	RegisteredCountry struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"registered_country"`
	GeolocationCountry struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	Continent struct {
		ISOCode string `maxminddb:"code"`
	} `maxminddb:"continent"`
}

func (c geoData) String() string {
	return fmt.Sprintf("registered country %s, geolocation country: %s, continent: %s", c.RegisteredCountry.ISOCode,
		c.GeolocationCountry.ISOCode, c.Continent.ISOCode)
}

func GeoUtilInitialise() {
	defer TimeTrack(time.Now(), "Initialising GeoDB")

	if !CheckPathExist(GEODB_FOLDER) {
		err := os.MkdirAll(GEODB_FOLDER, FILE_ACCESS_PERMISSION)
		Check(err)
	}
	geoFileFilePath := filepath.Join(GEODB_FOLDER, MAXMIND_GEO_DB_FILE_NAME)
	geoFileZipFilePath := filepath.Join(GEODB_FOLDER, MAXMIND_GEO_DB_ZIP_FILE_NAME)
	if !CheckPathExist(geoFileFilePath) {
		downloadMaxMindGeoLite()
		err := Ungzip(geoFileZipFilePath, GEODB_FOLDER)
		Check(err)
		maxMindCleanUp()
	} else {
		info, err := os.Stat(geoFileFilePath)
		Check(err)
		dateFile := time.Date(info.ModTime().Year(), info.ModTime().Month(), info.ModTime().Day(), 0, 0, 0, 0, time.UTC)
		dateNewFileAvailable := time.Date(time.Now().Year(), time.Now().Month(), firstTuesdayOfMonth(time.Now().Month()), 0, 0, 0, 0, time.UTC)

		if dateFile.Before(dateNewFileAvailable) && time.Now().After(dateNewFileAvailable) {
			os.Remove(geoFileFilePath)
			downloadMaxMindGeoLite()
			err := Ungzip(geoFileZipFilePath, GEODB_FOLDER)
			Check(err)
			maxMindCleanUp()
		}
	}

	var err error
	maxMindGeoDB, err = maxminddb.Open(geoFileFilePath)
	if err != nil {
		log.Fatal(err)
	}

}

func GeoUtilShutdown() {
	defer TimeTrack(time.Now(), "Shuting down GeoDB")
	maxMindGeoDB.Close()
}

func maxMindCleanUp() {
	geoFileZipFilePath := filepath.Join(GEODB_FOLDER, MAXMIND_GEO_DB_ZIP_FILE_NAME)
	os.Remove(geoFileZipFilePath)
	d, err := os.Open(GEODB_FOLDER)
	Check(err)

	defer d.Close()
	names, err := d.Readdirnames(-1)
	Check(err)

	for _, name := range names {
		if name != MAXMIND_GEO_DB_FILE_NAME {
			err := os.Remove(filepath.Join(GEODB_FOLDER, name))
			Check(err)
		}
	}
}

func FindGeoData(ip string) GeoData {
	if maxMindGeoDB == nil {
		panic(errors.New("GeoDB(s) have not been initialised! Initialise first."))
	}

	registeredCountryCode := "Not found"
	geolocationCountryCode := "Not found"
	continent := "Not found"

	ipToCheck := net.ParseIP(ip)
	if ipToCheck.To4() == nil {
		log.Print(errors.New(fmt.Sprintf("%v is not an IPv4 address\n", ipToCheck)))
	} else {
		var geoData geoData

		err := maxMindGeoDB.Lookup(ipToCheck, &geoData)
		if err != nil {
			log.Fatal(err)
		} else {
			registeredCountryCode = geoData.RegisteredCountry.ISOCode
			geolocationCountryCode = geoData.GeolocationCountry.ISOCode
			continent = geoData.Continent.ISOCode
		}
	}

	return GeoData{registeredCountryCode, geolocationCountryCode,
		continent}
}

func downloadMaxMindGeoLite() {
	geoFileZipFilePath := filepath.Join(GEODB_FOLDER, MAXMIND_GEO_DB_ZIP_FILE_NAME)

	out, err := os.Create(geoFileZipFilePath)
	defer out.Close()
	Check(err)

	resp, err1 := http.Get(MAXMIND_GEOIP_URL)
	defer resp.Body.Close()
	Check(err1)

	_, err2 := io.Copy(out, resp.Body)
	Check(err2)
}
