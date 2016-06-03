package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"github.com/agraphie/zversion/worker"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const NO_AGENT = "Not set"
const NO_AGENT_KEY = "Not set"
const ERROR_KEY = "Error"
const SERVER_AGENT_STRING = "Server:"
const SERVER_AGENT_DELIMITER = ":"
const OUTPUT_FILE_NAME = "http_version"
const FILE_ACCESS_PERMISSION = 0755
const MICROSOFT_IIS_AGENT_REGEX_STRING = `(?i)(Microsoft.IIS(?:\s|/)[0-9](?:\.[0-9]){0,2})`

type BaseEntry struct {
	IP        string
	Timestamp time.Time
	Error     string
}

type Entry struct {
	BaseEntry
	Data struct {
		Read string `json:",omitempty"`
	} `json:",omitempty"`
	Agent   string
	Error   string
	Version string
}

type HttpVersionResult struct {
	Started              time.Time
	Finished             time.Time
	ResultAmount         map[string]int
	CompleteResult       map[string][]Entry
	ProcessedZgrabOutput string
}

func (e Entry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Agent: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Agent)
}

func ParseHttpFile(path string) HttpVersionResult {
	fmt.Printf("Started at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	inputFileNameSplit := strings.Split(path, "/")
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	outputFile := util.CreateOutputJsonFile(util.ANALYSIS_OUTPUT_BASE_PATH+util.HTTP_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)

	httpVersionResult := HttpVersionResult{}
	httpVersionResult.Started = time.Now()

	hosts := worker.ParseFile(path, outputFile, workOnLine)
	httpVersionResult.ResultAmount = hosts.M
	httpVersionResult.Finished = time.Now()

	httpVersionResult.ProcessedZgrabOutput = path
	util.WriteSummaryFileAsJson(hosts.M, util.ANALYSIS_OUTPUT_BASE_PATH+util.HTTP_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)
	fmt.Printf("Finished at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	return httpVersionResult
}

func removeSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func workOnLine(queue chan string, complete chan bool, hosts *worker.HostsConcurrentSafe, writeQueue chan []byte) {
	for line := range queue {
		u := Entry{}
		json.Unmarshal([]byte(line), &u)
		dataAvailable := len(u.Data.Read) > 0
		contains := strings.Contains(u.Data.Read, SERVER_AGENT_STRING)

		if dataAvailable {
			u.Data.Read = strings.Replace(u.Data.Read, "\r\n", "\n", -1)
			u.Data.Read = strings.Replace(u.Data.Read, "\n\n", "\n", -1)
		}

		var key string

		//This caused a bug where "Internal Server Error" would also contain "Server" and thus this line
		//was assumed to contain the server version --> fixed to contain "Server:"
		switch {
		case u.Error != "":
			key = ERROR_KEY
		case u.Agent != "":
			key = u.Agent
		case dataAvailable && contains:
			splittedString := strings.Split(u.Data.Read, "\n")
			for i := range splittedString {
				if strings.Contains(splittedString[i], SERVER_AGENT_STRING) {
					serverSplit := strings.Split(splittedString[i], SERVER_AGENT_DELIMITER)
					if len(serverSplit) < 2 {
						log.Fatal(u.Data.Read)
					}
					agent := serverSplit[1]
					cleanAndAssign(agent, &u)
					//					u.Agent = removeSpaces(serverSplit[1])
					key = u.Agent + " " + u.Version

					break
				}
			}
		case dataAvailable && !contains:
			u.Agent = NO_AGENT
			key = NO_AGENT_KEY
		default:
			key = ERROR_KEY
		}
		u.Data.Read = ""
		worker.AddToMap(key, hosts)

		j, jerr := json.MarshalIndent(u, "", "  ")
		if jerr != nil {
			fmt.Println("jerr:", jerr.Error())
		}

		writeQueue <- j
	}
	complete <- true
}

func cleanAndAssign(agentString string, httpEntry *Entry) {
	microsoftIISRegex := regexp.MustCompile(MICROSOFT_IIS_AGENT_REGEX_STRING)

	IISMatch := microsoftIISRegex.FindStringSubmatch(agentString)

	if IISMatch != nil {
		httpEntry.Agent, httpEntry.Version = handleIISServer(IISMatch[1])

		return
	} else {
		httpEntry.Agent = agentString
	}
}

func handleIISServer(serverString string) (string, string) {
	split := strings.Split(serverString, "/")
	splitBlank := strings.Fields(serverString)

	server := "Microsoft-IIS"
	version := ""
	if len(split) > 1 {
		version = split[1]
	} else if len(splitBlank) > 1 {
		version = splitBlank[1]
	}
	if len(version) == 1 {
		version = version + ".0"
	}

	return server, version
}
