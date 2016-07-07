package util

import (
	"io"
	"net/http"
	"os"
)

var geoDB []geoLiteEntry

const GEODB_FOLDER = "geodb"
const MAXMIND_DB_ZIP_FILE_NAME = "GeoIPCountryCSV.zip"
const MAXMIND_DB_FILE_NAME = "GeoIPCountryWhois.csv"

const MAXMIND_GEOIP_URL = "http://geolite.maxmind.com/download/geoip/database/GeoIPCountryCSV.zip"

type geoLiteEntry struct {
	startIP     int
	endIp       int
	countryCode string
}

func GeoUtilInitialise() {
	if !CheckPathExist(GEODB_FOLDER) {
		err := os.MkdirAll(GEODB_FOLDER, FILE_ACCESS_PERMISSION)
		Check(err)
	}
	if !CheckPathExist(GEODB_FOLDER + "/" + MAXMIND_DB_FILE_NAME) {
		downloadMaxMindGeoLite()
		Unzip(GEODB_FOLDER+"/"+MAXMIND_DB_ZIP_FILE_NAME, GEODB_FOLDER)
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_DB_ZIP_FILE_NAME)
	} else if firstTuesdayOfMonth() {
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_DB_FILE_NAME)
		downloadMaxMindGeoLite()
		Unzip(GEODB_FOLDER+"/"+MAXMIND_DB_ZIP_FILE_NAME, GEODB_FOLDER)
		os.Remove(GEODB_FOLDER + "/" + MAXMIND_DB_ZIP_FILE_NAME)
	}
}

func findCountry(IP string) {

}

func downloadMaxMindGeoLite() {
	out, err := os.Create(GEODB_FOLDER + "/" + MAXMIND_DB_ZIP_FILE_NAME)
	defer out.Close()
	Check(err)

	resp, err1 := http.Get(MAXMIND_GEOIP_URL)
	defer resp.Body.Close()
	Check(err1)

	_, err2 := io.Copy(out, resp.Body)
	Check(err2)
}
