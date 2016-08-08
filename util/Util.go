package util

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const TIMESTAMP_FORMAT = "2006-01-02"
const TIMESTAMP_FORMAT_SECONDS = "2006-01-02T15:04:05-0700"

const ANALYSIS_OUTPUT_BASE_PATH = "cleanedResults"
const HTTP_ANALYSIS_OUTPUTH_PATH = "http"
const SSH_ANALYSIS_OUTPUTH_PATH = "ssh"
const HTTP_OUTPUT_FILE_NAME = "http_version"

const SCAN_OUTPUT_BASE_PATH = "scanResults"
const FILE_ACCESS_PERMISSION = 0755
const CONCURRENCY = 10000
const SCAN_CONCURRENCY = 1000

var HttpBaseOutputDir = filepath.Join(ANALYSIS_OUTPUT_BASE_PATH, HTTP_ANALYSIS_OUTPUTH_PATH)
var SSHBaseOutputDir = filepath.Join(ANALYSIS_OUTPUT_BASE_PATH, SSH_ANALYSIS_OUTPUTH_PATH)

type CleanMetaData struct {
	ServerHeaderCleaned    uint64    `json:"server_headers_cleaned"`
	ServerHeaderNotCleaned uint64    `json:"server_headers_not_cleaned"`
	Total                  uint64    `json:"total_processed"`
	InputFile              string    `json:"input_file"`
	Started                time.Time `json:"time_started"`
	Finished               time.Time `json:"time_finished"`
	Duration               string    `json:"duration"`
	Sha256SumOfInputFile   string    `json:"sha256_sum_of_input_file"`
	Sha256SumOfOutputFile  string    `json:"sha256_sum_of_output_file"`
	OutputFile             string    `json:"output_file"`
}

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
	filePath := filepath.Join(path, filename) + "_summary.json"
	f, err := os.Create(filePath)
	Check(err)
	defer f.Close()

	j, jerr := json.MarshalIndent(result, "", "  ")
	if jerr != nil {
		log.Println("jerr:", jerr.Error())
	}

	w := bufio.NewWriter(f)
	w.Write(j)
	w.Flush()
}

func CreateOutputJsonFile(path string, filename string) (*os.File, string) {
	if !CheckPathExist(path) {
		err := os.MkdirAll(path, FILE_ACCESS_PERMISSION)
		Check(err)
	}

	outputFileName := filepath.Join(path, filename) + ".json"
	f, err := os.Create(outputFileName)
	Check(err)

	return f, outputFileName
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

func WriteBytesToFile(wg *sync.WaitGroup, writeQueue chan []byte, file *os.File) {
	defer file.Close()

	w := bufio.NewWriter(file)

	for entry := range writeQueue {
		w.Write(entry)
		w.WriteString("\n")
	}

	w.Flush()

	wg.Done()
}

func WriteStringToFile(wg *sync.WaitGroup, writeQueue chan string, file *os.File) {
	defer file.Close()

	w := bufio.NewWriter(file)

	for entry := range writeQueue {
		w.WriteString(entry)
		w.WriteString("\n")
	}

	w.Flush()

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

func AppendZeroToVersion(version string) string {
	if len(version) == 1 {
		version = version + ".0"
	}

	return version
}

func firstTuesdayOfMonth() bool {
	result := false
	t := time.Now()

	if t.Day() <= 7 {
		if t.Weekday() == time.Tuesday {
			result = true
		}
	}

	return result
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, FILE_ACCESS_PERMISSION)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.FileInfo().Name())

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func Ungzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func Base64Decode(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func CalculateSha256(file string) string {
	isLz4 := regexp.MustCompile(`.*\.lz4`).FindStringSubmatch(file)
	isGz := regexp.MustCompile(`.*\.gz`).FindStringSubmatch(file)
	var result []byte
	var resultString string
	var err error
	if isLz4 != nil {
		cmd := "lz4 -dc " + file + " | sha256sum | awk '{print $1}'"
		result, err = exec.Command("bash", "-c", cmd).Output()
		resultString = strings.Replace(string(result), "\n", "", -1)
	} else if isGz != nil {
		cmd := "gunzip -dc " + file + " | sha256sum | awk '{print $1}'"
		result, err = exec.Command("bash", "-c", cmd).Output()
		resultString = strings.Replace(string(result), "\n", "", -1)
	} else {
		sha256 := sha256.New()
		file, err1 := os.Open(file)
		Check(err1)
		defer file.Close()
		_, errPipe := io.Copy(sha256, file)
		Check(errPipe)
		result = sha256.Sum(nil)
		resultString = hex.EncodeToString(result)
	}
	Check(err)
	return resultString
}
