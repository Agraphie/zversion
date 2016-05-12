package ssh

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
)

const NO_AGENT = "Not set"
const OUTPUT_FILE_NAME = "ssh_version"

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
	CompleteResult       []SSHEntry
	ResultAmount         map[string]int
	ProcessedZgrabOutput string
}

type hostsConcurrentSafe struct {
	sync.RWMutex
	m map[string]int
}

func (e SSHEntry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Software: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Software_version)
}

func ParseSSHFile(path string) SSHVersionResult {
	fmt.Printf("Started at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	inputFileNameSplit := strings.Split(path, "/")
	inputFileName := strings.Split(inputFileNameSplit[len(inputFileNameSplit)-1], ".")[0]
	outputFile := util.CreateOutputJsonFile(util.ANALYSIS_OUTPUT_BASE_PATH+util.SSH_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)

	sshVersionResult := SSHVersionResult{}
	sshVersionResult.Started = time.Now()
	hosts := parseFile(path, outputFile)

	sshVersionResult.ResultAmount = hosts.m
	sshVersionResult.Finished = time.Now()

	sshVersionResult.ProcessedZgrabOutput = path
	util.WriteSummaryFileAsJson(hosts.m, util.ANALYSIS_OUTPUT_BASE_PATH+util.SSH_ANALYSIS_OUTPUTH_PATH+inputFileName+"/", OUTPUT_FILE_NAME)
	fmt.Printf("Finished at %s\n", time.Now().Format(util.TIMESTAMP_FORMAT))

	return sshVersionResult
}

func addToMap(key string, hosts *hostsConcurrentSafe) {
	hosts.Lock()
	hosts.m["Total Processed"] = hosts.m["Total Processed"] + 1
	hosts.m[key] = hosts.m[key] + 1
	hosts.Unlock()
}

func parseFile(inputPath string, outputFile *os.File) hostsConcurrentSafe {
	var hostsResult = hostsConcurrentSafe{m: make(map[string]int)}
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
		go workOnLine(workQueue, complete, &hostsResult, writeQueue)
	}

	// Wait for everyone to finish.
	for i := 0; i < util.CONCURRENCY; i++ {
		<-complete
	}
	close(writeQueue)

	//wait for write queue
	<-complete

	return hostsResult
}

func workOnLine(queue chan string, complete chan bool, hosts *hostsConcurrentSafe, writeQueue chan []byte) {
	inputEntry := inputEntry{}

	for line := range queue {
		json.Unmarshal([]byte(line), &inputEntry)
		sshEntry := transform(inputEntry)

		if sshEntry.Software_version == "" {
			sshEntry.Software_version = NO_AGENT
		}
		key := sshEntry.Software_version

		addToMap(key, hosts)

		j, jerr := json.MarshalIndent(sshEntry, "", "  ")
		if jerr != nil {
			fmt.Println("jerr:", jerr.Error())
		}
		writeQueue <- j
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
