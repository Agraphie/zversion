package ssh

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"log"
	"os"
	"sync"
	"time"
	"strings"
)

const NO_AGENT = "Not set"
const NO_AGENT_KEY = "Not set"
const ERROR_KEY = "Error"
const SERVER_AGENT_STRING = "Server"
const SERVER_AGENT_DELIMITER = ":"
const OUTPUT_FILE_NAME = "ssh_version"
const OUTPUT_FILE_ENDING = ".json"
const FILE_ACCESS_PERMISSION = 0755

type BaseEntry struct {
	IP        string
	Timestamp time.Time
}

type SSHEntry struct {
	BaseEntry
	Raw_banner       string
	Protocol_version string
	Software_version string
	comments         string
}
type inputEntry struct {
	BaseEntry
	Data struct {
		SSH struct {
			Server_protocol struct {
				Raw_banner       string
				Protocol_version string
				Software_version string
				comments         string
			}
		}
	}
}

type SSHVersionResult struct {
	Started              time.Time
	Finished             time.Time
	CompleteResult       map[string][]SSHEntry
	ResultAmount         map[string]int
	ProcessedZgrabOutput string
}

type hostsConcurrentSafe struct {
	sync.RWMutex
	m map[string][]SSHEntry
}

func (e SSHEntry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Software: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Software_version)
}

func ParseSSHFile(path string) SSHVersionResult {
	fmt.Printf("Started at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	//
	sshVersionResult := SSHVersionResult{}
	sshVersionResult.Started = time.Now()
	hosts := parseFile(path)

	sshVersionResult.CompleteResult = hosts.m
	sshVersionResult.ResultAmount = sumUpResult(hosts)
	sshVersionResult.Finished = time.Now()
	inputFileNameSplit := strings.Split(path, "/")
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	sshVersionResult.ProcessedZgrabOutput = path

	writeMapToFile(util.ANALYSIS_OUTPUT_BASE_PATH+util.SSH_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME, sshVersionResult)
	fmt.Printf("Finished at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	return sshVersionResult
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

func addToMap(key string, entry SSHEntry, hosts *hostsConcurrentSafe) {
	hosts.Lock()
	hosts.m[key] = append(hosts.m[key], entry)
	fmt.Printf("Processed so far %d\n", len(hosts.m))

	hosts.Unlock()
}

func writeMapToFile(path string, filename string, sshVersionResult SSHVersionResult) {
	if !util.CheckPathExist(path) {
		err := os.MkdirAll(path, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}

	timestamp := time.Now().Format(util.TIMESTAMP_FORMAT)
	f, err := os.Create(path + filename + "_" + timestamp + OUTPUT_FILE_ENDING)
	util.Check(err)
	defer f.Close()

	j, jerr := json.MarshalIndent(sshVersionResult, "", "  ")
	if jerr != nil {
		fmt.Println("jerr:", jerr.Error())
	}

	w := bufio.NewWriter(f)
	w.Write(j)
	w.Flush()
}

var concurrency = 50

func parseFile(inputPath string) hostsConcurrentSafe {
	var hosts = hostsConcurrentSafe{m: make(map[string][]SSHEntry)}
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

//func progressLogging(complete chan bool, hosts *hostsConcurrentSafe, filepath string){
//
//	// Wait for everyone to finish.
//	for i := 0; i < concurrency; i++ {
//		for <- complete {
//			fmt.Println("Processed %d lines and %d% in total\n", len(hosts), )
//		}
//	}
//}

func workOnLine(queue chan string, complete chan bool, hosts *hostsConcurrentSafe) {
	for line := range queue {
		inputEntry := inputEntry{}
		json.Unmarshal([]byte(line), &inputEntry)
		sshEntry := transform(inputEntry)

		if sshEntry.Software_version == "" {
			sshEntry.Software_version = NO_AGENT
		}
		key := sshEntry.Software_version

		addToMap(key, sshEntry, hosts)
	}
	complete <- true
}

func transform(inputEntry inputEntry) SSHEntry {
	sshEntry := SSHEntry{
		Software_version: inputEntry.Data.SSH.Server_protocol.Software_version,
		Protocol_version: inputEntry.Data.SSH.Server_protocol.Protocol_version,
		comments:         inputEntry.Data.SSH.Server_protocol.comments,
		Raw_banner:       inputEntry.Data.SSH.Server_protocol.Raw_banner,
		BaseEntry:        inputEntry.BaseEntry}

	return sshEntry
}
