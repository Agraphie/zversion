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
const SERVER_FIELD_REGEXP_STRING = `(?:(?:\r\n)?Server:\s(.*)\r\n)`

var microsoftIISRegex = regexp.MustCompile(MICROSOFT_IIS_AGENT_REGEX_STRING)

type BaseEntry struct {
	IP        string
	Timestamp time.Time
	Error     string
}

type ZversionEntry struct {
	BaseEntry

	Agents []Server
	Error  string
}

type HttpVersionResult struct {
	Started              time.Time
	Finished             time.Time
	ResultAmount         map[string]int
	CompleteResult       map[string][]ZversionEntry
	ProcessedZgrabOutput string
}

func (e ZversionEntry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Agent: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Agents)
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
		u := RawZversionEntry{}
		json.Unmarshal([]byte(line), &u)
		httpEntry := ZversionEntry{BaseEntry: u.BaseEntry, Error: u.Error}

		serverFields := serverFieldRegexp.FindAllStringSubmatch(u.Data.Read, -1)

		//This caused a bug where "Internal Server Error" would also contain "Server" and thus this line
		//was assumed to contain the server version --> fixed to contain "Server:"
		switch {
		case httpEntry.Error != "":
			worker.AddToMap(ERROR_KEY, hosts)
		case len(serverFields) > 0:
			for _, v := range serverFields {
				server := v[1]
				cleanAndAssign(server, &httpEntry)
			}
			for _, v := range httpEntry.Agents {
				var key string
				if v.Version != "" {
					key = v.Agent + " " + v.Version
				} else {
					key = v.Agent
				}
				worker.AddToMap(key, hosts)
			}
		case len(u.Data.Read) > 0 && len(serverFields) == 0:
			httpEntry.Agents = append(httpEntry.Agents, Server{Agent: NO_AGENT})
			worker.AddToMap(NO_AGENT, hosts)
		default:
			worker.AddToMap(ERROR_KEY, hosts)
		}

		j, jerr := json.MarshalIndent(httpEntry, "", "  ")
		if jerr != nil {
			fmt.Println("jerr:", jerr.Error())
		}

		writeQueue <- j
	}
	complete <- true
}

func cleanAndAssign(agentString string, httpEntry *ZversionEntry) {
	IISMatch := microsoftIISRegex.FindStringSubmatch(agentString)

	if IISMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleIISServer(IISMatch))
		return
	} else {
		httpEntry.Agents = append(httpEntry.Agents, Server{Agent: agentString})
	}
}

func handleIISServer(serverString []string) Server {
	server := "Microsoft-IIS"
	version := serverString[1]
	if len(version) == 1 {
		version = version + ".0"
	} else if len(version) > 5 {
		fmt.Println(serverString)
	}

	return Server{Agent: server, Version: version}
}
