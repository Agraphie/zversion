package ssh

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/analysis"
	"github.com/agraphie/zversion/util"
	"github.com/agraphie/zversion/worker"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

const OUTPUT_FILE_NAME = "ssh_version"
const ERROR_KEY = "Error"
const SSH_VERSION_INVALID = "SSH protocol version invalid"
const SSH_CLEANING_META_DATA_FILE_NAME = "ssh_meta_data.json"

type BaseEntry struct {
	IP        string
	Timestamp time.Time
}

type SSHEntry struct {
	BaseEntry
	Raw_banner       string
	Protocol_version string
	Comments         string
	Vendor           string
	SoftwareVersion  string
	CanonicalVersion string
	Error            string
	GeoData          util.GeoData
	ASId             string
	ASOwner          string
}

type inputEntry struct {
	BaseEntry
	Data struct {
		SSH struct {
			Server_protocol struct {
				Raw_banner       string
				Protocol_version string
				Software_version string
				comments         string
			}
		}
	}
	Error string
}

type SSHVersionResult struct {
	Started              time.Time
	Finished             time.Time
	CompleteResult       []SSHEntry
	ResultAmount         map[string]int
	ProcessedZgrabOutput string
}

type sshCleanMetaData struct {
	ServerHeaderCleaned    uint64    `json:"server_headers_cleaned"`
	ServerHeaderNotCleaned uint64    `json:"server_headers_not_cleaned"`
	Total                  uint64    `json:"total_processed"`
	InputFile              string    `json:"input_file"`
	Started                time.Time `json:"time_started"`
	Finished               time.Time `json:"time_finished"`
	Duration               string    `json:"duration"`
}

func (e SSHEntry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Vendor: %v, Software version: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Vendor, e.SoftwareVersion)
}

var totalProcessed uint64 = 0

func ParseSSHFile(path string) SSHVersionResult {
	defer util.TimeTrack(time.Now(), "Processing")
	metaDate := sshCleanMetaData{}
	metaDate.Started = time.Now()
	metaDate.InputFile = path
	log.Println("Start cleaning...")

	inputFileNameSplit := strings.Split(path, string(filepath.Separator))
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	outputFolderPath := filepath.Join(util.ANALYSIS_OUTPUT_BASE_PATH, util.SSH_ANALYSIS_OUTPUTH_PATH, inputFileName)
	outputFile := util.CreateOutputJsonFile(outputFolderPath, OUTPUT_FILE_NAME)

	sshVersionResult := SSHVersionResult{}
	sshVersionResult.Started = time.Now()
	util.GeoUtilInitialise()
	util.ASUtilInitialise()

	hosts := worker.ParseFile(path, outputFile, workOnLine)

	sshVersionResult.ResultAmount = hosts.M
	sshVersionResult.Finished = time.Now()

	sshVersionResult.ProcessedZgrabOutput = path
	util.WriteSummaryFileAsJson(hosts.M, outputFolderPath, OUTPUT_FILE_NAME)
	log.Println("Cleaning finished")
	log.Printf("Not cleaned: %d\n", softwareBannerNotCleaned)

	log.Println("Start analysis...")
	analysis.RunSSHAnalyseScripts(filepath.Join(outputFolderPath, OUTPUT_FILE_NAME+".json"), outputFolderPath, nil)
	log.Println("Analysis finished")

	metaDate.Duration = time.Since(metaDate.Started).String()

	metaDate.Finished = time.Now()
	metaDate.ServerHeaderCleaned = softwareBannerCleaned
	metaDate.ServerHeaderNotCleaned = softwareBannerNotCleaned
	metaDate.Total = totalProcessed
	writeMetDataToFile(metaDate, outputFolderPath)

	return sshVersionResult
}

func workOnLine(queue chan string, complete chan bool, hosts *worker.HostsConcurrentSafe, writeQueue chan []byte) {
	inputEntry := inputEntry{}
	sshRegexp := regexp.MustCompile(`.*SSH.*`)

	for line := range queue {
		json.Unmarshal([]byte(line), &inputEntry)
		sshEntry := transform(inputEntry)
		//Assign country
		sshEntry.GeoData = util.FindGeoData(sshEntry.IP)

		//Assign AS
		sshEntry.ASId, sshEntry.ASOwner = util.FindAS(sshEntry.IP)
		var key string
		if sshRegexp.FindStringSubmatch(sshEntry.Raw_banner) != nil && sshEntry.Raw_banner != "" {
			if inputEntry.Data.SSH.Server_protocol.Software_version != "" {
				cleanAndAssign(inputEntry.Data.SSH.Server_protocol.Software_version, &sshEntry)
			} else {
				cleanAndAssign(sshEntry.Raw_banner, &sshEntry)
			}

			if sshEntry.SoftwareVersion != "" {
				key = sshEntry.Vendor + " " + sshEntry.SoftwareVersion
			} else {
				key = sshEntry.Vendor
			}
		} else {
			key = ERROR_KEY
		}

		worker.AddToMap(key, hosts)

		j, jerr := json.Marshal(sshEntry)
		if jerr != nil {
			fmt.Println("jerr:", jerr.Error())
		}
		atomic.AddUint64(&totalProcessed, 1)

		writeQueue <- j
	}
	complete <- true
}

func transform(inputEntry inputEntry) SSHEntry {
	sshEntry := SSHEntry{
		Protocol_version: inputEntry.Data.SSH.Server_protocol.Protocol_version,
		Comments:         inputEntry.Data.SSH.Server_protocol.comments,
		Raw_banner:       inputEntry.Data.SSH.Server_protocol.Raw_banner,
		BaseEntry:        inputEntry.BaseEntry,
		Error:            inputEntry.Error}

	return sshEntry
}

func writeMetDataToFile(output sshCleanMetaData, outputPath string) {
	filePath := filepath.Join(outputPath, SSH_CLEANING_META_DATA_FILE_NAME)
	f, err := os.Create(filePath)
	util.Check(err)
	defer f.Close()

	j, jerr := json.Marshal(output)
	util.Check(jerr)

	w := bufio.NewWriter(f)
	w.Write(j)
	w.Flush()
}
