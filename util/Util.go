package util

import "os"
const TIMESTAMP_FORMAT = "2006-01-02-15:04:00"

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
	panic(e)
}