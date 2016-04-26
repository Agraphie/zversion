package util

import "os"
const TIMESTAMP_FORMAT = "2006-01-02-15:04:00"

const ANALYSIS_OUTPUT_BASE_PATH = "analysisResults/"
const HTTP_ANALYSIS_OUTPUTH_PATH = "http/"
const SCAN_OUTPUT_BASE_PATH = "scanResults/"

func CheckPathExist(path string) bool{
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}


func Check(e error) {
	if e != nil {
		panic(e)
	}
}