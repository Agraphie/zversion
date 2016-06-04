package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"github.com/agraphie/zversion/worker"
	"regexp"
	"strings"
	"time"
)

const NO_AGENT = "No server field in header"
const NO_AGENT_KEY = "No server field in header"
const ERROR_KEY = "Error"
const OUTPUT_FILE_NAME = "http_version"
const FILE_ACCESS_PERMISSION = 0755
const MICROSOFT_IIS_AGENT_REGEX_STRING = `(?i)(?:Microsoft.IIS(?:(?:\s|/)([0-9](?:\.[0-9]){0,2})){0,1})`
const SERVER_FIELD_REGEXP_STRING = `(?:\r\nServer:\s(.*)\r\n)`

var microsoftIISRegex = regexp.MustCompile(MICROSOFT_IIS_AGENT_REGEX_STRING)

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

func workOnLine(queue chan string, complete chan bool, hosts *worker.HostsConcurrentSafe, writeQueue chan []byte) {
	serverFieldRegexp := regexp.MustCompile(SERVER_FIELD_REGEXP_STRING)
	for line := range queue {
		u := Entry{}
		json.Unmarshal([]byte(line), &u)

		serverFields := serverFieldRegexp.FindAllStringSubmatch(u.Data.Read, -1)

		var key string

		//This caused a bug where "Internal Server Error" would also contain "Server" and thus this line
		//was assumed to contain the server version --> fixed to contain "Server:"
		switch {
		case u.Error != "":
			key = ERROR_KEY
		case u.Agent != "":
			key = u.Agent
		case len(serverFields) > 0:
			for _, v := range serverFields {
				server := v[1]
				cleanAndAssign(server, &u)
				if u.Version != "" {
					key = u.Agent + " " + u.Version
				} else {
					key = u.Agent
				}

				break
			}
		case len(u.Data.Read) > 0 && len(serverFields) == 0:
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
	IISMatch := microsoftIISRegex.FindStringSubmatch(agentString)

	if IISMatch != nil {
		httpEntry.Agent, httpEntry.Version = handleIISServer(IISMatch)

		return
	} else {
		httpEntry.Agent = agentString
	}
}

func handleIISServer(serverString []string) (string, string) {
	server := "Microsoft-IIS"
	version := serverString[1]
	if len(version) == 1 {
		version = version + ".0"
	} else if len(version) > 5 {
		fmt.Println(serverString)
	}

	return server, version
}
