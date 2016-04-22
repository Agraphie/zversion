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
)

const NO_AGENT = "Not set"
const NO_AGENT_KEY = "Not set"
const ERROR_KEY = "Error"
const SERVER_AGENT_STRING = "Server"
const SERVER_AGENT_DELIMITER = ":"
const OUTPUT_FILE_LOCATION = "http_logs"
const OUTPUT_FILE_NAME = "http_version"
const OUTPUT_FILE_ENDING = ".json"
const FILE_ACCESS_PERMISSION = 0755

const TIMESTAMP_FORMAT = "2006-01-02-15:04:00"


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
}

var hosts = struct {
	sync.RWMutex
	m map[string][]Entry
}{m: make(map[string][]Entry)}

func (e Entry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Agent: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Agent)
}

func ParseHttpFile(path string) HttpVersionResult {
	httpVersionResult := HttpVersionResult{}
	httpVersionResult.Started = time.Now()

	worker.ParseFile(path, workOnLine)

	httpVersionResult.CompleteResult = hosts.m
	httpVersionResult.ResultAmount = sumUpResult()
	httpVersionResult.Finished = time.Now()

	writeMapToFile(OUTPUT_FILE_LOCATION + "/", OUTPUT_FILE_NAME, httpVersionResult)

	return httpVersionResult
}

func sumUpResult() map[string]int {
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

func workOnLine(queue chan string) {
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
		addToMap(key, u)
	}
}

func addToMap(key string, entry Entry) {
	hosts.Lock()
	hosts.m[key] = append(hosts.m[key], entry)
	hosts.Unlock()
}

func writeMapToFile(path string, filename string, httpVersionResult HttpVersionResult) {
	if !checkPathExist() {
		err := os.MkdirAll(OUTPUT_FILE_LOCATION, FILE_ACCESS_PERMISSION)
		check(err)
	}

	timestamp := time.Now().Format(TIMESTAMP_FORMAT)
	f, err := os.Create(path + filename + "_" + timestamp + OUTPUT_FILE_ENDING)
	check(err)
	defer f.Close()

	j, jerr := json.MarshalIndent(httpVersionResult, "", "  ")
	if jerr != nil {
		fmt.Println("jerr:", jerr.Error())
	}

	w := bufio.NewWriter(f)
	w.Write(j)
	w.Flush()
}

func checkPathExist() bool {
	_, err := os.Stat(OUTPUT_FILE_LOCATION)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func check(e error) {
	if e != nil {
		if os.IsNotExist(e) {
			os.Mkdir("http_logs", 770)
		} else {
			panic(e)
		}
	}
}