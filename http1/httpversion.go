package http1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

const NO_AGENT = "Not set"
const NO_AGENT_KEY = "Not set"
const ERROR_KEY = "Error"
const SERVER_AGENT_STRING = "Server:"
const SERVER_AGENT_DELIMITER = ":"
const OUTPUT_FILE_NAME = "http_version"
const OUTPUT_FILE_ENDING = ".json"
const FILE_ACCESS_PERMISSION = 0755

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
	Agent string
	Error string
}

type HttpVersionResult struct {
	Started              time.Time
	Finished             time.Time
	ResultAmount         map[string]int
	CompleteResult       map[string][]Entry
	ProcessedZgrabOutput string
}

type hostsConcurrentSafe struct {
	sync.RWMutex
	m map[string]int
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
	hosts := parseFile(path, outputFile)

	httpVersionResult.ResultAmount = hosts.m
	httpVersionResult.Finished = time.Now()

	httpVersionResult.ProcessedZgrabOutput = path
	util.WriteSummaryFileAsJson(hosts.m, util.ANALYSIS_OUTPUT_BASE_PATH+util.HTTP_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)
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

func addToMap(key string, hosts *hostsConcurrentSafe) {
	hosts.Lock()
	hosts.m["Total Processed"] = hosts.m["Total Processed"] + 1
	hosts.m[key] = hosts.m[key] + 1
	hosts.Unlock()
}

func parseFile(inputPath string, outputFile *os.File) hostsConcurrentSafe {
	var hosts = hostsConcurrentSafe{m: make(map[string]int)}
	// This channel has no buffer, so it only accepts input when something is ready
	// to take it out. This keeps the reading from getting ahead of the writers.
	workQueue := make(chan string)
	writeQueue := make(chan []byte)

	// We need to know when everyone is done so we can exit.
	complete := make(chan bool)

	// Read the lines into the work queue.
	go func() {
		file, err := os.Open(inputPath)
		if err != nil {
			log.Fatal(err)
		}

		// Close when the functin returns
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			workQueue <- scanner.Text()
		}

		// Close the channel so everyone reading from it knows we're done.
		close(workQueue)
	}()

	//start writer
	go util.WriteEntries(complete, writeQueue, outputFile)

	// Now read them all off, concurrently.
	for i := 0; i < util.CONCURRENCY; i++ {
		go workOnLine(workQueue, complete, &hosts, writeQueue)
	}

	// Wait for everyone to finish.
	for i := 0; i < util.CONCURRENCY; i++ {
		<-complete
	}

	close(writeQueue)

	//wait for write queue
	<-complete

	return hosts
}

func workOnLine(queue chan string, complete chan bool, hosts *hostsConcurrentSafe, writeQueue chan []byte) {
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
		case dataAvailable && contains:
			splittedString := strings.Split(u.Data.Read, "\n")
			for i := range splittedString {
				if strings.Contains(splittedString[i], SERVER_AGENT_STRING) {
					serverSplit := strings.Split(splittedString[i], SERVER_AGENT_DELIMITER)
					if len(serverSplit) < 2 {
						log.Fatal(u.Data.Read)
					}
					u.Agent = removeSpaces(serverSplit[1])
					key = u.Agent

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
		addToMap(key, hosts)

		j, jerr := json.MarshalIndent(u, "", "  ")
		if jerr != nil {
			fmt.Println("jerr:", jerr.Error())
		}

		writeQueue <- j
	}
	complete <- true
}
