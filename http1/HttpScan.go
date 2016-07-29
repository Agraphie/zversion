package http1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	HTTP_SCAN_OUTPUTH_PATH         = "http"
	HTTP_SCAN_DEFAULT_PORT         = "80"
	HTTP_SCAN_DEFAULT_SCAN_TARGETS = "10000"
	HTTP_SCAN_DEFAULT_PPS          = "100000"
	SCAN_OUTPUT_FILE_NAME_FULL     = "zversion_full"
	SCAN_OUTPUT_FILE_NAME_VHOST    = "zversion_vhost"
	SCAN_ZGRAB_ERROR_FILE_NAME     = "zgrab_error"
	META_DATA_FILE_NAME            = "scan_meta_data"
)

type RunningHttpScan struct {
	RunningCommands []*exec.Cmd
	ProgressZmap    float64
	ProgressZgrab   float32
	Started         time.Time
	Finished        time.Time
}

type RawZversionEntry struct {
	BaseEntry
	Data struct {
		Read string `json:",omitempty"`
	} `json:",omitempty"`
	Error string
	Body  string
}

type MetaData struct {
	Port          int     `json:"port"`
	Success_count uint32  `json:"success_count"`
	Failure_count uint32  `json:"failure_count"`
	Total         int     `json:"total"`
	Start_time    string  `json:"start_time"`
	End_time      string  `json:"end_time"`
	Duration      int     `json:"duration"`
	Senders       int     `json:"senders"`
	Timeout       int     `json:"timeout"`
	Tls_version   *string `json:"tls_version"`
	Mail_type     *string `json:"mail_type"`
	Ca_file_name  *string `json:"ca_file_name"`
	Sni_support   bool    `json:"sni_support"`
	ZgrabRequest  string  `json:"zgrab_request"`
	FallbackCount uint32  `json:"fallback_count"`
	ScanInputFile *string `json:"input_file"`
}

var zgrabRequest string
var fallbackCount uint32
var metaDataString string
var zmapInputFile *string
var isVHostScan bool

const TIMEOUT_IN_SECONDS = "60"
const TIMEOUT_IN_SECONDS_INT = 60
const MAX_KB_TO_READ = "64"

/**
commands is a map where the key is the timestamp when the scan was launched and the values are all cmds which are
running for that timestamp. This makes it easier to kill them off.
*/
func LaunchHttpScan(runningScan *RunningHttpScan, scanOutputPath string, port string, scanTargets string, blacklistFile string, inputFile string) {
	started := time.Now()
	timestampFormatted := started.Format(util.TIMESTAMP_FORMAT)

	outputPath := filepath.Join(scanOutputPath, HTTP_SCAN_OUTPUTH_PATH, timestampFormatted)
	if !util.CheckPathExist(outputPath) {
		err := os.MkdirAll(outputPath, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}

	if inputFile == "" {
		zmapInputFile = nil
		launchFullHttpScan(timestampFormatted, outputPath, port, scanTargets, blacklistFile)
	} else {
		zmapInputFile = &inputFile
		isVHostScan = checkVHostScan(inputFile)
		launchRestrictedHttpScan(outputPath, timestampFormatted, port, inputFile)
	}
	log.Printf("Http scan done in: %s\n", time.Since(started))
}

func launchRestrictedHttpScan(outputPath string, timestampFormatted string, port string, inputFile string) {
	var c3 *exec.Cmd
	if isVHostScan {
		c3 = exec.Command("zgrab", "--port", port, "--data=./http-req-domain", "--timeout", TIMEOUT_IN_SECONDS, "--input-file", inputFile, "--http-max-size", MAX_KB_TO_READ)
		content, _ := ioutil.ReadFile("./http-req-domain")
		zgrabRequest = string(content)
	} else {
		c3 = exec.Command("zgrab", "--port", port, "--data=./http-req", "--timeout", TIMEOUT_IN_SECONDS, "--input-file", inputFile, "--http-max-size", MAX_KB_TO_READ)
		content, _ := ioutil.ReadFile("./http-req")
		zgrabRequest = string(content)
	}

	c3StdOut, _ := c3.StdoutPipe()
	var wg sync.WaitGroup
	wg.Add(1)
	go handleZgrabOutput(outputPath, timestampFormatted, c3StdOut, &wg)

	_ = c3.Start()
	wg.Wait()

	_ = c3.Wait()
}

func checkVHostScan(inputFile string) bool {
	file, err := os.Open(inputFile)
	util.Check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	firstLine := scanner.Text()

	util.Check(scanner.Err())

	splitComma := strings.Split(firstLine, ",")

	return len(splitComma) == 2
}

func launchFullHttpScan(timestampFormatted string, outputPath string, port string, scanTargets string, blacklistFile string) {
	nmapOutputFileName := "zmap_output_" + timestampFormatted + ".csv"

	zmapErrorLog := "zmap_error_" + timestampFormatted
	zmapErr, _ := os.Create(filepath.Join(outputPath, zmapErrorLog))

	defer zmapErr.Close()

	var c1 *exec.Cmd
	if blacklistFile == "null" {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS)
	} else {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS, "-b", blacklistFile)
	}

	c2 := exec.Command("ztee", filepath.Join(outputPath, nmapOutputFileName))
	c3 := exec.Command("zgrab", "--port", port, "--data=./http-req", "--timeout", TIMEOUT_IN_SECONDS, "--http-max-size", MAX_KB_TO_READ)
	content, _ := ioutil.ReadFile("./http-req")
	zgrabRequest = string(content)
	//if runningScan != nil {
	//	runningScan.RunningCommands = append(runningScan.RunningCommands, c1)
	//	runningScan.RunningCommands = append(runningScan.RunningCommands, c2)
	//	runningScan.RunningCommands = append(runningScan.RunningCommands, c3)
	//	if runningScan.Started.IsZero() {
	//		runningScan.Started = started
	//	}
	//}

	c3StdOut, _ := c3.StdoutPipe()

	c2.Stderr = os.Stderr

	c2.Stdin, _ = c1.StdoutPipe()
	c3.Stdin, _ = c2.StdoutPipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go handleZgrabOutput(outputPath, timestampFormatted, c3StdOut, &wg)
	c1.Stderr = io.MultiWriter(zmapErr, os.Stderr)

	_ = c3.Start()
	_ = c2.Start()
	_ = c1.Run()

	_ = c2.Wait()
	wg.Wait()
	_ = c3.Wait()
	//c3StdOut.Close()

	//finished := time.Now()
	//if runningScan != nil {
	//	runningScan.Finished = finished
	//}
}

func handleZgrabOutput(currentScanPath string, timestampFormatted string, stdOut io.ReadCloser, wg *sync.WaitGroup) {
	var zgrabOutputFileName string
	if isVHostScan {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_VHOST + "_" + timestampFormatted + ".json"
	} else {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_FULL + "_" + timestampFormatted + ".json"
	}

	zgrabErrorLog := SCAN_ZGRAB_ERROR_FILE_NAME + "_" + timestampFormatted
	metaDataFileName := META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
	metaDataFile, _ := os.Create(filepath.Join(currentScanPath, metaDataFileName))
	zgrabErr, _ := os.Create(filepath.Join(currentScanPath, zgrabErrorLog))
	zgrabOut, _ := os.Create(filepath.Join(currentScanPath, zgrabOutputFileName))

	writeQueueErr := make(chan []byte, 100000)
	writeQueueOut := make(chan string, 100000)
	writeQueueMetaData := make(chan []byte, 2000)

	var wgWriters sync.WaitGroup
	wgWriters.Add(3)
	go util.WriteBytesToFile(&wgWriters, writeQueueErr, zgrabErr)
	go util.WriteStringToFile(&wgWriters, writeQueueOut, zgrabOut)
	go util.WriteBytesToFile(&wgWriters, writeQueueMetaData, metaDataFile)

	stdOutScanner := bufio.NewScanner(stdOut)

	var wgWorkers sync.WaitGroup
	wgWorkers.Add(util.CONCURRENCY)
	workQueue := make(chan string, 100000)
	//start workers
	for i := 0; i < util.CONCURRENCY; i++ {
		go workOnZgrabOutputLine(workQueue, &wgWorkers, writeQueueErr, writeQueueOut)
	}

	for stdOutScanner.Scan() {
		workQueue <- stdOutScanner.Text()
	}

	close(workQueue)
	wgWorkers.Wait()
	writeMetaData(metaDataString, writeQueueMetaData)

	close(writeQueueOut)
	close(writeQueueErr)
	close(writeQueueMetaData)
	wgWriters.Wait()

	wg.Done()

}

func workOnZgrabOutputLine(workQueue chan string, wg *sync.WaitGroup, writeQueueErr chan []byte, writeQueueOut chan string) {
	for line := range workQueue {
		if strings.Contains(line, "success_count") {
			metaDataString = line
			continue
		}

		u := RawZversionEntry{}
		json.Unmarshal([]byte(line), &u)
		if u.Error != "" {
			handleZgrabError(u, writeQueueOut, writeQueueErr)
		} else {
			writeQueueOut <- line
		}
	}

	wg.Done()
}

func writeMetaData(line string, writeQueueMetaData chan []byte) {
	var metaData MetaData
	json.Unmarshal([]byte(line), &metaData)
	metaData.ZgrabRequest = zgrabRequest
	metaData.FallbackCount = fallbackCount
	metaData.Success_count += fallbackCount
	metaData.Failure_count -= fallbackCount
	metaData.ScanInputFile = zmapInputFile

	j, _ := json.Marshal(metaData)

	writeQueueMetaData <- j
	os.Stdout.WriteString(string(j) + "\n")
}

func handleZgrabError(entry RawZversionEntry, outFile chan string, errFile chan []byte) {
	timeout := time.Duration(TIMEOUT_IN_SECONDS_INT * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	response, err := client.Get("http://" + entry.BaseEntry.IP)
	if err == nil {
		for _, v := range response.Header["Server"] {
			entry.Data.Read += "\r\n" + "Server: " + v
		}
		if entry.Data.Read != "" {
			entry.Data.Read += "\r\n"
		}
		entry.Error = ""
		if response.Body != nil {
			defer response.Body.Close()
			bs := make([]byte, 64000)
			response.Body.Read(bs)
			//		bs, err := ioutil.ReadAll(response.Body)
			if err != nil {
				entry.Error = err.Error()
			} else {
				entry.Body = string(bs)
			}
		}
		atomic.AddUint32(&fallbackCount, 1)
	} else {
		errFile <- []byte(entry.BaseEntry.IP + ": " + entry.Error)
	}

	j, _ := json.Marshal(entry)
	outFile <- string(j)
}

func progressZgrab(zmapStdOut io.ReadCloser, zgrabStdOut io.ReadCloser, runningScan *RunningHttpScan) {
	zmapScanner := bufio.NewScanner(zmapStdOut)
	zgrabScanner := bufio.NewScanner(zgrabStdOut)

	zmapLinesProcessed := float32(1.0)
	zgrabLinesProcessed := float32(0)

	go func() {
		for zmapScanner.Scan() {
			zmapLinesProcessed++

		}
	}()

	for zgrabScanner.Scan() {
		zgrabLinesProcessed++

		runningScan.ProgressZgrab = zgrabLinesProcessed / zmapLinesProcessed * 100
		fmt.Println(zgrabScanner.Text())

	}
}

func progressAndLogZmap(reader io.ReadCloser, logWriter io.Writer, runningScan *RunningHttpScan) {
	in := bufio.NewScanner(reader)

	for in.Scan() {
		logWriter.Write(in.Bytes())
		logWriter.Write([]byte("\n"))
		progress := strings.Fields(in.Text())[2]
		progressNoPercent := strings.Split(progress, "%")[0]
		runningScan.ProgressZmap, _ = strconv.ParseFloat(progressNoPercent, 64)
	}
}
