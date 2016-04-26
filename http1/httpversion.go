package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/worker"
	"sync"
	"strings"
	"unicode"
	"bufio"
	"os"
	"time"
	"github.com/agraphie/zversion/util"
	"log"
)

const NO_AGENT = "Not set"
const NO_AGENT_KEY = "Not set"
const ERROR_KEY = "Error"
const SERVER_AGENT_STRING = "Server"
const SERVER_AGENT_DELIMITER = ":"
const OUTPUT_FILE_NAME = "http_version"
const OUTPUT_FILE_ENDING = ".json"
const FILE_ACCESS_PERMISSION = 0755



type Entry struct {
	worker.BaseEntry
	Data  struct {
			  Read  string
			  Write string
		  }
	Agent string
	Error string
}

type HttpVersionResult struct {
	Started time.Time
	Finished time.Time
	CompleteResult map[string][]Entry
	ResultAmount   map[string]int
	ProcessedZgrabOutput string
}

type hostsConcurrentSafe struct {
	sync.RWMutex
	m map[string][]Entry
}

func (e Entry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Agent: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Agent)
}

func ParseHttpFile(path string) HttpVersionResult {
	fmt.Println(path)
	httpVersionResult := HttpVersionResult{}
	httpVersionResult.Started = time.Now()

	hosts := parseFile(path)

	httpVersionResult.CompleteResult = hosts.m
	httpVersionResult.ResultAmount = sumUpResult(hosts)
	httpVersionResult.Finished = time.Now()
	inputFileNameSplit := strings.Split(path, "/")
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	writeMapToFile(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH + inputFileName + "/", OUTPUT_FILE_NAME, httpVersionResult)
	httpVersionResult.ProcessedZgrabOutput = path

	return httpVersionResult
}

func sumUpResult(hosts hostsConcurrentSafe) map[string]int {
	summedUp := make(map[string]int)
	for key, _ := range hosts.m {
		for range hosts.m[key] {
			summedUp[key] = summedUp[key] + 1
		}
	}

	return summedUp
}
func removeSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}



func addToMap(key string, entry Entry, hosts *hostsConcurrentSafe) {
	hosts.Lock()
	hosts.m[key] = append(hosts.m[key], entry)
	hosts.Unlock()
}

func writeMapToFile(path string, filename string, httpVersionResult HttpVersionResult) {
	if !util.CheckPathExist(path) {
		err := os.MkdirAll(path, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}

	timestamp := time.Now().Format(util.TIMESTAMP_FORMAT)
	f, err := os.Create(path + filename + "_" + timestamp + OUTPUT_FILE_ENDING)
	util.Check(err)
	defer f.Close()

	j, jerr := json.MarshalIndent(httpVersionResult, "", "  ")
	if jerr != nil {
		fmt.Println("jerr:", jerr.Error())
	}

	w := bufio.NewWriter(f)
	w.Write(j)
	w.Flush()
}


type BaseEntry struct {
	IP        string
	Timestamp time.Time
	Error     string
}

var concurrency = 100

func parseFile(inputPath string) hostsConcurrentSafe{
	var hosts = hostsConcurrentSafe{m: make(map[string][]Entry)}
	// This channel has no buffer, so it only accepts input when something is ready
	// to take it out. This keeps the reading from getting ahead of the writers.
	workQueue := make(chan string)

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

	// Now read them all off, concurrently.
	for i := 0; i < concurrency; i++ {
		go workOnLine(workQueue, complete, &hosts)
	}

	// Wait for everyone to finish.
	for i := 0; i < concurrency; i++ {
		<-complete
	}
	return hosts
}

func workOnLine(queue chan string, complete chan bool, hosts *hostsConcurrentSafe) {
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

		switch {
		case dataAvailable && contains:
			splittedString := strings.Split(u.Data.Read, "\n")
			for i := range splittedString {
				if strings.Contains(splittedString[i], SERVER_AGENT_STRING) {
					serverSplit := strings.Split(splittedString[i], SERVER_AGENT_DELIMITER)
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
		addToMap(key, u, hosts)
	}
	complete <- true
}
