package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"github.com/agraphie/zversion/worker"
	"strings"
	"time"
)

const NO_AGENT = "No server field in header"
const NO_AGENT_KEY = "No server field in header"
const ERROR_KEY = "Error"
const OUTPUT_FILE_NAME = "http_version"
const FILE_ACCESS_PERMISSION = 0755

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

type RawCensysEntry struct {
	BaseEntry

	Data struct {
		Http struct {
			Response struct {
				Headers struct {
					Server []string
				}
			}
		}
		Error string
	}
}
type RawCensysLegacyEntry struct {
	BaseEntry
	Data struct {
		Http struct {
			Response struct {
				Headers struct {
					Server *string `json:",omitempty"`
				}
			}
		}
		Error string
	}
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
	for line := range queue {
		rawCensysEntry := RawCensysEntry{}
		json.Unmarshal([]byte(line), &rawCensysEntry)
		rawCensysLegacyEntry := RawCensysLegacyEntry{}
		json.Unmarshal([]byte(line), &rawCensysLegacyEntry)

		u := RawZversionEntry{}

		if len(rawCensysEntry.Data.Http.Response.Headers.Server) != 0 || rawCensysEntry.Data.Error != "" {
			u.BaseEntry = rawCensysEntry.BaseEntry
			u.Error = rawCensysEntry.Error
			for _, v := range rawCensysEntry.Data.Http.Response.Headers.Server {
				u.Data.Read += "\r\n" + "Server: " + v
			}
			if u.Data.Read != "" {
				u.Data.Read += "\r\n"
			}
		} else if rawCensysLegacyEntry.Data.Http.Response.Headers.Server != nil || rawCensysLegacyEntry.Data.Error != "" {
			u.BaseEntry = rawCensysLegacyEntry.BaseEntry
			u.Error = rawCensysLegacyEntry.Error
			u.Data.Read += "\r\n" + "Server: " + *rawCensysLegacyEntry.Data.Http.Response.Headers.Server
			if u.Data.Read != "" {
				u.Data.Read += "\r\n"
			}
		} else {
			json.Unmarshal([]byte(line), &u)
		}

		httpEntry := ZversionEntry{BaseEntry: u.BaseEntry, Error: u.Error}
		serverFields := serverFieldRegexp.FindAllStringSubmatch(u.Data.Read, -1)

		//This caused a bug where "Internal Server Error" would also contain "Server" and thus this line
		//was assumed to contain the server version --> fixed to contain "Server:"
		switch {
		case httpEntry.Error != "":
			httpEntry.Agents = make([]Server, 0)
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
		case (len(u.Data.Read) > 0 && len(serverFields) == 0) || u.Error == "":
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
