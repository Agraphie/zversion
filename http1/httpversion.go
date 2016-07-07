package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"github.com/agraphie/zversion/worker"
	"log"
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
type Server struct {
	Vendor           string
	Version          string
	CanonicalVersion string
}
type CMS struct {
	Vendor           string
	Version          string
	CanonicalVersion string
}

type ZversionEntry struct {
	BaseEntry

	Agents  []Server
	Error   string
	CMS     []CMS
	Country string
	ASId    string
	ASOwner string
}

type unknownHeaderField struct {
	Key   string
	Value []string
}

type RawCensysEntry struct {
	BaseEntry

	Data struct {
		Http struct {
			Response struct {
				Headers struct {
					Server  []string
					Unknown []unknownHeaderField
				}
				Body string
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
					Server  *string `json:",omitempty"`
					Unknown []unknownHeaderField
				}
				Body string
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
	util.GeoUtilInitialise()
	util.ASUtilInitialise()
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
		u, xContentHeaderField := assignRawEntry(line)
		httpEntry := ZversionEntry{BaseEntry: u.BaseEntry, Error: u.Error}
		serverFields := serverFieldRegexp.FindAllStringSubmatch(u.Data.Read, -1)

		//assign CMS if available
		cleanAndAssignCMS(u.Body, xContentHeaderField, &httpEntry)
		//Assign country
		httpEntry.Country = util.FindCountry(httpEntry.IP)

		//Assign AS
		httpEntry.ASId, httpEntry.ASOwner = util.FindAS(httpEntry.IP)
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
					key = v.Vendor + " " + v.Version
				} else {
					key = v.Vendor
				}
				worker.AddToMap(key, hosts)
			}
		case (len(u.Data.Read) > 0 && len(serverFields) == 0) || u.Error == "":
			httpEntry.Agents = append(httpEntry.Agents, Server{Vendor: NO_AGENT})
			worker.AddToMap(NO_AGENT, hosts)
		default:
			worker.AddToMap(ERROR_KEY, hosts)
		}

		j, jerr := json.Marshal(httpEntry)
		if jerr != nil {
			fmt.Println("jerr:", jerr.Error())
		}

		if len(writeQueue) == cap(writeQueue) {
			log.Println("Write queue is full! Blocking routine.")
		}

		writeQueue <- j
	}
	complete <- true
}

func assignRawEntry(rawLine string) (RawZversionEntry, []string) {
	rawCensysEntry := RawCensysEntry{}
	json.Unmarshal([]byte(rawLine), &rawCensysEntry)
	rawCensysLegacyEntry := RawCensysLegacyEntry{}
	json.Unmarshal([]byte(rawLine), &rawCensysLegacyEntry)
	u := RawZversionEntry{}
	var xContentEncodedBy []string

	if len(rawCensysEntry.Data.Http.Response.Headers.Server) != 0 || rawCensysEntry.Data.Error != "" {
		u.BaseEntry = rawCensysEntry.BaseEntry
		u.Error = rawCensysEntry.Error
		u.Body = rawCensysEntry.Data.Http.Response.Body

		for _, v := range rawCensysEntry.Data.Http.Response.Headers.Server {
			u.Data.Read += "\r\n" + "Server: " + v
		}
		if u.Data.Read != "" {
			u.Data.Read += "\r\n"
		}
		xContentEncodedBy = findXContentEncodedByField(rawCensysEntry.Data.Http.Response.Headers.Unknown)
	} else if rawCensysLegacyEntry.Data.Http.Response.Headers.Server != nil || rawCensysLegacyEntry.Data.Error != "" {
		u.BaseEntry = rawCensysLegacyEntry.BaseEntry
		u.Error = rawCensysLegacyEntry.Error
		u.Body = rawCensysLegacyEntry.Data.Http.Response.Body
		u.Data.Read += "\r\n" + "Server: " + *rawCensysLegacyEntry.Data.Http.Response.Headers.Server
		if u.Data.Read != "" {
			u.Data.Read += "\r\n"
		}
		xContentEncodedBy = findXContentEncodedByField(rawCensysEntry.Data.Http.Response.Headers.Unknown)
	} else {
		//TODO: does not have an x_content_encoded_by header field!
		json.Unmarshal([]byte(rawLine), &u)
	}

	return u, xContentEncodedBy
}

func findXContentEncodedByField(unknownHeaders []unknownHeaderField) []string {
	var xContentEncodedBy []string
	for _, v := range unknownHeaders {
		if v.Key == "x_content_encoded_by" {
			xContentEncodedBy = v.Value
		}
	}

	return xContentEncodedBy
}
