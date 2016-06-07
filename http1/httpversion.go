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
const MICROSOFT_IIS_SERVER_REGEX_STRING = `(?i)(?:Microsoft.IIS(?:(?:\s|/)(\d+(?:\.\d){0,2})){0,1})`
const APACHE_SERVER_REGEX_STRING = `(?i)(?:Apache(?:(?:\s|/)(\d+(?:\.\d+){0,2}(?:-(?:M|B)\d)?)){0,1})`

const BASE_REGEX = `(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,2})){0,1})`
const LIGHTHTTPD_SERVER_REGEX_STRING = `(?i)(?:Lighthttpd` + BASE_REGEX
const NGINX_SERVER_REGEX_STRING = `(?i)(?:nginx` + BASE_REGEX
const ATS_SERVER_REGEX_STRING = `(?i)(?:ATS` + BASE_REGEX
const BOA_SERVER_REGEX_STRING = `(?i)(?:boa(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,2}(?:(?:rc)\d+)?)){0,1})`

const SERVER_FIELD_REGEXP_STRING = `(?:(?:\r\n)Server:\s(.*)\r\n)`

var microsoftIISRegex = regexp.MustCompile(MICROSOFT_IIS_SERVER_REGEX_STRING)
var apacheRegex = regexp.MustCompile(APACHE_SERVER_REGEX_STRING)
var nginxRegex = regexp.MustCompile(NGINX_SERVER_REGEX_STRING)
var lighthttpdRegex = regexp.MustCompile(LIGHTHTTPD_SERVER_REGEX_STRING)
var atsRegex = regexp.MustCompile(ATS_SERVER_REGEX_STRING)
var boaRegex = regexp.MustCompile(BOA_SERVER_REGEX_STRING)

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
	fmt.Printf("Not cleaned: %d\n", notCleaned)

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
			//first get the server fields and clean them
			for _, v := range serverFields {
				server := v[1]
				cleanAndAssign(server, &httpEntry)
			}
			//then increase the count for all included servers
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

var notCleaned = 0

func cleanAndAssign(agentString string, httpEntry *ZversionEntry) {
	IISMatch := microsoftIISRegex.FindStringSubmatch(agentString)
	if IISMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleIISServer(IISMatch))
		return
	}

	apacheMatch := apacheRegex.FindStringSubmatch(agentString)
	if apacheMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleApacheServer(apacheMatch))
		return
	}

	nginxMatch := nginxRegex.FindStringSubmatch(agentString)
	if nginxMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleNginxServer(nginxMatch))
		return
	}

	lighthttpdMatch := lighthttpdRegex.FindStringSubmatch(agentString)
	if lighthttpdMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleLighthttpdServer(lighthttpdMatch))
		return
	}

	atsMatch := atsRegex.FindStringSubmatch(agentString)
	if atsMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleATSServer(atsMatch))
		return
	}

	boaMatch := boaRegex.FindStringSubmatch(agentString)
	if boaMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleBOAServer(boaMatch))
		return
	}

	httpEntry.Agents = append(httpEntry.Agents, Server{Agent: agentString})
	notCleaned++
}

func handleBOAServer(serverString []string) Server {
	server := "BOA"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleATSServer(serverString []string) Server {
	server := "ATS"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleLighthttpdServer(serverString []string) Server {
	server := "lighthttpd"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleIISServer(serverString []string) Server {
	server := "Microsoft-IIS"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}
func handleApacheServer(serverString []string) Server {
	server := "Apache"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}
func handleNginxServer(serverString []string) Server {
	server := "nginx"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func appendZero(version string) string {
	if len(version) == 1 {
		version = version + ".0"
	}

	return version
}
