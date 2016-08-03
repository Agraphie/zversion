package http1

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
	"strings"
	"sync/atomic"
	"time"
)

type EntryType string

const (
	NO_AGENT                                    = "No server field in header"
	NO_AGENT_KEY                                = "No server field in header"
	ERROR_KEY                                   = "Error"
	FILE_ACCESS_PERMISSION                      = 0755
	LEGACY_CENSYS                     EntryType = "LegacyCensys"
	CENSYS                            EntryType = "Censys"
	RAPID7                            EntryType = "Rapid7"
	ZVERSION                          EntryType = "Zversion"
	HTTP_CLEANING_META_DATA_FILE_NAME           = "http_meta_data.json"
)

type BaseEntry struct {
	IP             string
	Domain         string
	Timestamp      time.Time
	Error          string
	InputEntryType EntryType
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
	GeoData util.GeoData
	ASId    string
	ASOwner string
}

type unknownHeaderField struct {
	Key   string
	Value []string
}

type RawRapid7Entry struct {
	BaseEntry

	VHost string
	Host  string
	Data  string
}

type httpCleanMetaData struct {
	ServerHeaderCleaned    uint64    `json:"server_headers_cleaned"`
	ServerHeaderNotCleaned uint64    `json:"server_headers_not_cleaned"`
	CMSCleaned             uint64    `json:"cms_cleaned"`
	Total                  uint64    `json:"total_processed"`
	InputFile              string    `json:"input_file"`
	Started                time.Time `json:"time_started"`
	Finished               time.Time `json:"time_finished"`
	Duration               string    `json:"duration"`
	Sha256SumOfFile        string    `json:"sha256_sum_of_input_file"`
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

var totalProcessed uint64 = 0

func (e ZversionEntry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Agent: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Agents)
}

func ParseHttpFile(path string) HttpVersionResult {
	defer util.TimeTrack(time.Now(), "Processing")
	metaDate := httpCleanMetaData{}
	metaDate.Started = time.Now()
	metaDate.InputFile = path
	log.Println("Start cleaning...")

	inputFileNameSplit := strings.Split(path, string(filepath.Separator))
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	outputFolderPath := filepath.Join(util.ANALYSIS_OUTPUT_BASE_PATH, util.HTTP_ANALYSIS_OUTPUTH_PATH, inputFileName)
	outputFile := util.CreateOutputJsonFile(outputFolderPath, util.HTTP_OUTPUT_FILE_NAME)

	httpVersionResult := HttpVersionResult{}
	httpVersionResult.Started = time.Now()
	util.GeoUtilInitialise()
	util.ASUtilInitialise()
	hosts := worker.ParseFile(path, outputFile, workOnLine)
	util.GeoUtilShutdown()
	httpVersionResult.ResultAmount = hosts.M
	httpVersionResult.Finished = time.Now()

	httpVersionResult.ProcessedZgrabOutput = path
	util.WriteSummaryFileAsJson(hosts.M, outputFolderPath, util.HTTP_OUTPUT_FILE_NAME)
	log.Println("Cleaning finished")
	log.Printf("Not cleaned: %d\n", serverHeaderNotCleaned)

	log.Println("Start analysis...")
	analysis.RunHTTPAnalyseScripts(filepath.Join(outputFolderPath, util.HTTP_OUTPUT_FILE_NAME+".json"), outputFolderPath, nil)
	log.Println("Analysis finished")
	metaDate.Duration = time.Since(metaDate.Started).String()

	metaDate.Finished = time.Now()
	metaDate.ServerHeaderCleaned = serverHeaderCleaned
	metaDate.ServerHeaderNotCleaned = serverHeaderNotCleaned
	metaDate.CMSCleaned = cmsCleaned
	metaDate.Total = totalProcessed

	metaDate.Sha256SumOfFile = util.CalculateSha256(path)
	writeMetDataToFile(metaDate, outputFolderPath)

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
		httpEntry.GeoData = util.FindGeoData(httpEntry.IP)

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
		atomic.AddUint64(&totalProcessed, 1)
		writeQueue <- j
	}
	complete <- true
}

func assignRawEntry(rawLine string) (RawZversionEntry, []string) {
	rawCensysEntry := RawCensysEntry{}
	rawCensysLegacyEntry := RawCensysLegacyEntry{}
	rawRapid7Entry := RawRapid7Entry{}
	json.Unmarshal([]byte(rawLine), &rawCensysEntry)
	json.Unmarshal([]byte(rawLine), &rawCensysLegacyEntry)
	json.Unmarshal([]byte(rawLine), &rawRapid7Entry)

	u := RawZversionEntry{}
	var xContentEncodedBy []string

	if rawRapid7Entry.Data != "" || rawRapid7Entry.Host != "" {
		u.BaseEntry = rawRapid7Entry.BaseEntry
		u.BaseEntry.InputEntryType = RAPID7

		var err error
		u.Data.Read, err = util.Base64Decode(rawRapid7Entry.Data)
		if err != nil {
			log.Printf("Error for IP %s", u.BaseEntry.IP)
			log.Println(err)
		}
	} else if len(rawCensysEntry.Data.Http.Response.Headers.Server) != 0 || rawCensysEntry.Data.Error != "" {
		u.BaseEntry = rawCensysEntry.BaseEntry
		u.BaseEntry.InputEntryType = CENSYS
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
		u.BaseEntry.InputEntryType = LEGACY_CENSYS
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
		u.BaseEntry.InputEntryType = ZVERSION
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

func writeMetDataToFile(output httpCleanMetaData, outputPath string) {
	filePath := filepath.Join(outputPath, HTTP_CLEANING_META_DATA_FILE_NAME)
	f, err := os.Create(filePath)
	util.Check(err)
	defer f.Close()

	j, jerr := json.Marshal(output)
	util.Check(jerr)

	w := bufio.NewWriter(f)
	w.Write(j)
	w.WriteString("\n")
	w.Flush()
}
