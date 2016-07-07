package ssh

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"github.com/agraphie/zversion/worker"
	"regexp"
	"strings"
	"time"
)

const OUTPUT_FILE_NAME = "ssh_version"
const ERROR_KEY = "Error"
const SSH_VERSION_INVALID = "SSH protocol version invalid"

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
	Country          string
	AS               string
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

func (e SSHEntry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Vendor: %v, Software version: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Vendor, e.SoftwareVersion)
}

func ParseSSHFile(path string) SSHVersionResult {
	fmt.Printf("Started at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	inputFileNameSplit := strings.Split(path, "/")
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	outputFile := util.CreateOutputJsonFile(util.ANALYSIS_OUTPUT_BASE_PATH+util.SSH_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)

	sshVersionResult := SSHVersionResult{}
	sshVersionResult.Started = time.Now()
	util.GeoUtilInitialise()

	hosts := worker.ParseFile(path, outputFile, workOnLine)

	sshVersionResult.ResultAmount = hosts.M
	sshVersionResult.Finished = time.Now()

	sshVersionResult.ProcessedZgrabOutput = path
	util.WriteSummaryFileAsJson(hosts.M, util.ANALYSIS_OUTPUT_BASE_PATH+util.SSH_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)
	fmt.Printf("Finished at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	return sshVersionResult
}

func workOnLine(queue chan string, complete chan bool, hosts *worker.HostsConcurrentSafe, writeQueue chan []byte) {
	inputEntry := inputEntry{}
	sshRegexp := regexp.MustCompile(`.*SSH.*`)

	for line := range queue {
		json.Unmarshal([]byte(line), &inputEntry)
		sshEntry := transform(inputEntry)
		//Assign country
		sshEntry.Country = util.FindCountry(sshEntry.IP)

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
