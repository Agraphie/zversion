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
	} else if firstTuesdayOfMonth() {
		os.Remove(geoFileFilePath)
		downloadMaxMindGeoLite()
		err := Ungzip(geoFileZipFilePath, GEODB_FOLDER)
		Check(err)
		maxMindCleanUp()
	}
	//readInMaxMindGeoDBCSV()
	//sort.Sort(Asc(maxMindGeoDB))
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

//func FindGeoData(ip string) GeoData {
//	if len(maxMindGeoDB) == 0 {
//		panic(errors.New("GeoDB(s) have not been initialised! Initialise first."))
//	}
//	registeredCountryCode := "Not found"
//	geolocationCountryCode := "Not found"
//	registeredContinentCode := "Not found"
//	geolocationContinentCode := "Not found"
//
//	ipToCheck := net.ParseIP(ip)
//	if ipToCheck.To4() == nil {
//		panic(errors.New(fmt.Sprintf("%v is not an IPv4 address\n", ipToCheck)))
//	}
//	for _, v := range maxMindGeoDB {
//		firstIP := net.ParseIP(v.firstIP)
//		lastIP := net.ParseIP(v.lastIP)
//
//		if bytes.Compare(ipToCheck, firstIP) >= 0 && bytes.Compare(ipToCheck, lastIP) <= 0 {
//			registeredCountryCode = v.registeredCountryCode
//			geolocationCountryCode = v.geolocationCountryCode
//			registeredContinentCode = v.registeredContinentCode
//			geolocationContinentCode = v.geolocationContinentCode
//			break
//		}
//
//		//	if v.network.Contains(ipToCheck) {
//		//		registeredCountryCode = v.registeredCountryCode
//		//		geolocationCountryCode = v.geolocationCountryCode
//		//		registeredContinentCode = v.registeredContinentCode
//		//		geolocationContinentCode = v.geolocationContinentCode
//		//		break
//		//	}
//	}
//
//	return GeoData{registeredCountryCode, geolocationCountryCode,
//		registeredContinentCode, geolocationContinentCode}
//}
//
//func readInMaxMindGeoDBCSV() {
//	readInCountryCodes()
//	geoFileFilePath := filepath.Join(GEODB_FOLDER, MAXMIND_GEO_DB_FILE_NAME)
//
//	dbCsv, err := os.Open(geoFileFilePath)
//	defer dbCsv.Close()
//	Check(err)
//
//	reader := csv.NewReader(bufio.NewReader(dbCsv))
//	//Skip header line
//	reader.Read()
//	for {
//		record, err := reader.Read()
//		// Stop at EOF.
//		if err == io.EOF {
//			break
//		}
//		firstIP, network, errIp := net.ParseCIDR(record[0])
//		Check(errIp)
//		lastIP := getLastIP(firstIP, network)
//
//		registeredCountryId := record[1]
//		representedCountryId := record[2]
//
//		entry := geoLiteEntry{network: network, registeredCountryCode: countries[registeredCountryId].countryCode, geolocationCountryCode: countries[representedCountryId].countryCode,
//			registeredContinentCode: countries[registeredCountryId].continentCode, geolocationContinentCode: countries[representedCountryId].continentCode, firstIP: firstIP.String(), lastIP: lastIP.String()}
//		maxMindGeoDB = append(maxMindGeoDB, entry)
//	}
//}
func getLastIP(firstIP net.IP, network *net.IPNet) net.IP {
	result := make(net.IP, 4)
	for ip := firstIP.Mask(network.Mask); network.Contains(ip); inc(ip) {
		copy(result, ip)
	}
	return result
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

//func readInCountryCodes() {
//	geoFileCountryFilePath := filepath.Join(GEODB_FOLDER, MAXMIND_GEO_DB_COUNTRY_CODES_FILE_NAME)
//	dbCodesCsv, err1 := os.Open(geoFileCountryFilePath)
//	defer dbCodesCsv.Close()
//	Check(err1)
//	reader := csv.NewReader(bufio.NewReader(dbCodesCsv))
//	countries = make(map[string]countryEntry)
//
//	for {
//		record, err := reader.Read()
//		// Stop at EOF.
//		if err == io.EOF {
//			break
//		}
//		country := countryEntry{countryCode: record[4], continentCode: record[2]}
//		countries[record[0]] = country
//	}
//}

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
