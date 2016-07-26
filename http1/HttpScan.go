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
	SCAN_OUTPUT_FILE_NAME          = "zversion_full"
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
		launchRestrictedHttpScan(outputPath, timestampFormatted, port, inputFile)
	}
	log.Printf("Http scan done in: %s\n", time.Since(started))
}

func launchRestrictedHttpScan(outputPath string, timestampFormatted string, port string, inputFile string) {
	c3 := exec.Command("zgrab", "--port", port, "--data=./http-req-head", "--input-file", inputFile)
	c3StdOut, _ := c3.StdoutPipe()
	c3StdErr, _ := c3.StderrPipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go handleZgrabOutput(outputPath, timestampFormatted, c3StdOut, c3StdErr, &wg)

	_ = c3.Start()
	_ = c3.Wait()

	c3StdOut.Close()
	c3StdErr.Close()
	wg.Wait()
}

func launchFullHttpScan(timestampFormatted string, outputPath string, port string, scanTargets string, blacklistFile string) {
	nmapOutputFileName := "zmap_output_" + timestampFormatted + ".csv"

	zmapErrorLog := "zmap_error_" + timestampFormatted
	zmapErr, _ := os.Create(filepath.Join(outputPath, zmapErrorLog))
	zmapErrW := io.WriteCloser(zmapErr)

	//defer zmapErrW.Flush()
	defer zmapErr.Close()

	var c1 *exec.Cmd
	if blacklistFile == "null" {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS)
	} else {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS, "-b", blacklistFile)
	}

	c2 := exec.Command("ztee", filepath.Join(outputPath, nmapOutputFileName))
	c3 := exec.Command("zgrab", "--port", port, "--data=./http-req-head")
	//if runningScan != nil {
	//	runningScan.RunningCommands = append(runningScan.RunningCommands, c1)
	//	runningScan.RunningCommands = append(runningScan.RunningCommands, c2)
	//	runningScan.RunningCommands = append(runningScan.RunningCommands, c3)
	//	if runningScan.Started.IsZero() {
	//		runningScan.Started = started
	//	}
	//}

	c1StdErr, _ := c1.StderrPipe()
	c3StdOut, _ := c3.StdoutPipe()
	c3StdErr, _ := c3.StderrPipe()

	c2.Stderr = os.Stderr

	//	c3.Stderr = zgrabErrW

	c2.Stdin, _ = c1.StdoutPipe()
	c3.Stdin, _ = c2.StdoutPipe()
	//	c3.Stdout = os.Stdout
	var wg sync.WaitGroup
	wg.Add(2)
	go handleZgrabOutput(outputPath, timestampFormatted, c3StdOut, c3StdErr, &wg)

	_ = c2.Start()
	_ = c3.Start()
	_ = c1.Start()

	go printAndLog(c1StdErr, zmapErrW, &wg)

	_ = c2.Wait()
	_ = c3.Wait()
	_ = c1.Wait()

	c1StdErr.Close()
	c3StdOut.Close()
	c3StdErr.Close()
	wg.Wait()

	//finished := time.Now()
	//if runningScan != nil {
	//	runningScan.Finished = finished
	//}

}

func handleZgrabOutput(currentScanPath string, timestampFormatted string, stdOut io.ReadCloser, stdErr io.ReadCloser, wg *sync.WaitGroup) {
	zgrabOutputFileName := SCAN_OUTPUT_FILE_NAME + "_" + timestampFormatted + ".json"
	zgrabErrorLog := SCAN_ZGRAB_ERROR_FILE_NAME + "_" + timestampFormatted
	metaDataFileName := META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
	metaDataFile, _ := os.Create(filepath.Join(currentScanPath, metaDataFileName))
	zgrabErr, _ := os.Create(filepath.Join(currentScanPath, zgrabErrorLog))
	zgrabOut, _ := os.Create(filepath.Join(currentScanPath, zgrabOutputFileName))

	writeQueueErr := make(chan string)
	writeQueueOut := make(chan string)
	writeQueueMetaData := make(chan string)

	var wgWriters sync.WaitGroup
	wgWriters.Add(3)
	go util.WriteStringToFile(&wgWriters, writeQueueErr, zgrabErr)
	go util.WriteStringToFile(&wgWriters, writeQueueOut, zgrabOut)
	go util.WriteStringToFile(&wgWriters, writeQueueMetaData, metaDataFile)

	stdOutScanner := bufio.NewScanner(stdOut)

	var wgWorkers sync.WaitGroup
	wgWorkers.Add(util.CONCURRENCY)
	workQueue := make(chan []byte)
	//start workers
	for i := 0; i < util.CONCURRENCY; i++ {
		go workOnZgrabOutputLine(workQueue, &wgWorkers, writeQueueErr, writeQueueOut)
	}

	for stdOutScanner.Scan() {
		workQueue <- stdOutScanner.Bytes()
	}

	close(workQueue)
	wgWorkers.Wait()
	writeMetaData(metaDataString, writeQueueMetaData)

	close(writeQueueOut)
	close(writeQueueErr)
	close(writeQueueMetaData)
	wgWriters.Wait()

	zgrabErr.Close()
	zgrabOut.Close()
	metaDataFile.Close()
	wg.Done()

}

func workOnZgrabOutputLine(workQueue chan []byte, wg *sync.WaitGroup, writeQueueErr chan string, writeQueueOut chan string) {
	for line := range workQueue {
		lineString := string(line)
		if strings.Contains(lineString, "success_count") {
			metaDataString = lineString
			continue
		}

		u := RawZversionEntry{}
		json.Unmarshal(line, &u)
		if u.Error != "" {
			handleZgrabError(u, writeQueueOut, writeQueueErr)
		} else {
			writeQueueOut <- lineString + "\n"
		}
	}

	wg.Done()
}

func writeMetaData(line string, writeQueueMetaData chan string) {
	var metaData MetaData
	json.Unmarshal([]byte(line), &metaData)
	metaData.ZgrabRequest = zgrabRequest
	metaData.FallbackCount = fallbackCount
	metaData.Success_count += fallbackCount
	metaData.Failure_count -= fallbackCount
	metaData.ScanInputFile = zmapInputFile

	j, _ := json.Marshal(metaData)

	writeQueueMetaData <- string(j) + "\n"
	os.Stdout.WriteString(string(j) + "\n")
}

func handleZgrabError(entry RawZversionEntry, outFile chan string, errFile chan string) {
	timeout := time.Duration(10 * time.Second)
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
			bs, err := ioutil.ReadAll(response.Body)
			util.Check(err)
			entry.Body = string(bs)
		}
		atomic.AddUint32(&fallbackCount, 1)
	} else {
		errFile <- entry.BaseEntry.IP + ": " + entry.Error + "\n"
	}

	j, _ := json.Marshal(entry)
	outFile <- string(j) + "\n"
}

func printAndLog(reader io.ReadCloser, logWriter io.Writer, wg *sync.WaitGroup) {
	in := bufio.NewScanner(reader)

	for in.Scan() {
		logWriter.Write(in.Bytes())
		logWriter.Write([]byte("\n"))

		fmt.Println(in.Text())
	}
	wg.Done()
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
