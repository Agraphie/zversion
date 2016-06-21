package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

const TIMESTAMP_FORMAT = "2006-01-02-15:04:05"

const ANALYSIS_OUTPUT_BASE_PATH = "analysisResults/"
const HTTP_ANALYSIS_OUTPUTH_PATH = "http/"
const SSH_ANALYSIS_OUTPUTH_PATH = "ssh/"

const SCAN_OUTPUT_BASE_PATH = "scanResults/"
const FILE_ACCESS_PERMISSION = 0755
const CONCURRENCY = 8000

func CheckPathExist(path string) bool {
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

func WriteSummaryFileAsJson(result map[string]int, path string, filename string) {
	filePath := path + filename + "_summary.json"
	f, err := os.Create(filePath)
	Check(err)
	defer f.Close()

	j, jerr := json.MarshalIndent(result, "", "  ")
	if jerr != nil {
		fmt.Println("jerr:", jerr.Error())
	}

	w := bufio.NewWriter(f)
	w.Write(j)
	w.Flush()
}

func CreateOutputJsonFile(path string, filename string) *os.File {
	if !CheckPathExist(path) {
		err := os.MkdirAll(path, FILE_ACCESS_PERMISSION)
		Check(err)
	}

	f, err := os.Create(path + filename + ".json")
	Check(err)

	return f
}

func WriteEntries(complete chan bool, writeQueue chan []byte, file *os.File) {
	defer file.Close()

	w := bufio.NewWriter(file)
	//w.WriteString("{\n[\n")

	for entry := range writeQueue {
		w.Write(entry)
		w.WriteString("\n")
	}

	//w.WriteString("\n]\n}\n")

	w.Flush()

	complete <- true
}

func WriteStringToFile(wg *sync.WaitGroup, writeQueue chan string, file *os.File) {
	defer file.Close()

	w := io.WriteCloser(file)

	for entry := range writeQueue {
		w.Write([]byte(entry))
	}

	wg.Done()
}

func MakeVersionCanonical(version string) string {
	canonicalVersion := ""
	numbersExtract := regexp.MustCompile(`\d*`)
	splitVersion := strings.Split(version, ".")

	for _, v := range splitVersion {
		currentVersion := numbersExtract.FindStringSubmatch(v)[0]

		switch len(currentVersion) {
		case 1:
			canonicalVersion = canonicalVersion + "000" + currentVersion
		case 2:
			canonicalVersion = canonicalVersion + "00" + currentVersion
		case 3:
			canonicalVersion = canonicalVersion + "0" + currentVersion
		case 4:
			canonicalVersion = canonicalVersion + currentVersion
		}
	}

	switch len(canonicalVersion) {
	case 4:
		canonicalVersion = canonicalVersion + "000000000000"
	case 8:
		canonicalVersion = canonicalVersion + "00000000"
	case 12:
		canonicalVersion = canonicalVersion + "0000"
	}

	return string(canonicalVersion)
}
