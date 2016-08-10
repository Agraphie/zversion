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
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	HTTP_SCAN_OUTPUTH_PATH           = "http"
	HTTP_SCAN_DEFAULT_PORT           = "80"
	HTTP_SCAN_DEFAULT_SCAN_TARGETS   = "10000"
	HTTP_SCAN_DEFAULT_PPS            = "100000"
	SCAN_OUTPUT_FILE_NAME_FULL       = "zversion_full"
	SCAN_OUTPUT_FILE_NAME_VHOST      = "zversion_vhost"
	SCAN_ZGRAB_ERROR_FILE_NAME       = "zgrab_error"
	ZVERSION_META_DATA_FILE_NAME     = "zversion_scan_meta_data"
	ZMAP_META_DATA_FILE_NAME         = "zmap_scan_meta_data.json"
	ZVERSION_SUCCESSFUL_FALLBACK_IPS = "zversion_scan_successful_fallback_ips"
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
	Port             int     `json:"port"`
	Success_count    uint32  `json:"success_count"`
	Failure_count    uint32  `json:"failure_count"`
	Total            int     `json:"total"`
	Start_time       string  `json:"start_time"`
	End_time         string  `json:"end_time"`
	Duration         int     `json:"duration"`
	Senders          int     `json:"senders"`
	Timeout          int     `json:"timeout"`
	Tls_version      *string `json:"tls_version"`
	Mail_type        *string `json:"mail_type"`
	Ca_file_name     *string `json:"ca_file_name"`
	Sni_support      bool    `json:"sni_support"`
	ZgrabRequest     string  `json:"zgrab_request"`
	ScanInputFile    *string `json:"input_file"`
	ScanOutputFile   string  `json:"output_file"`
	Sha256OutputFile string  `json:"sha256_sum_of_output_file"`
}

var zgrabRequest string
var fallbackCount uint32
var metaDataString string
var zmapInputFile *string
var isVHostScan bool

const TIMEOUT_IN_SECONDS_FIRST_TRY = "10"
const TIMEOUT_IN_SECONDS_FIRST_TRY_INT = 10

const TIMEOUT_IN_SECONDS_SECOND_TRY = "15"
const TIMEOUT_IN_SECONDS_SECOND_TRY_INT = 15
const MAX_KB_TO_READ = "64"

/**
commands is a map where the key is the timestamp when the scan was launched and the values are all cmds which are
running for that timestamp. This makes it easier to kill them off.
*/
func LaunchHttpScan(runningScan *RunningHttpScan, scanOutputPath string, port string, scanTargets string, blacklistFile string, inputFile string) {
	started := time.Now()
	timestampFormatted := started.Format(util.TIMESTAMP_FORMAT_SECONDS)

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
	var cmdScanData string
	if isVHostScan {
		cmdScanData = "--data=./http-req-domain"
		content, _ := ioutil.ReadFile("./http-req-domain")
		zgrabRequest = string(content)
	} else {
		cmdScanData = "--data=./http-req-domain"
		content, _ := ioutil.ReadFile("./http-req")
		zgrabRequest = string(content)
	}

	outputFile := filepath.Join(outputPath, getZgrabOutputFilename(timestampFormatted))
	metaDataFileName := ZVERSION_META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
	cmdScanString := "zgrab " + cmdScanData + " --senders 2500 --timeout " + TIMEOUT_IN_SECONDS_FIRST_TRY + " --input-file " + inputFile + " --output-file=" + outputFile + " --metadata-file=" + filepath.Join(outputPath, metaDataFileName)
	scanCmd := exec.Command("bash", "-c", cmdScanString)
	scanCmd.Stderr = os.Stderr
	var wg sync.WaitGroup
	wg.Add(1)
	runErr := scanCmd.Run()
	util.Check(runErr)
	wg.Wait()

	metaDataFile := filepath.Join(outputPath, metaDataFileName)
	enhanceMetaData(metaDataFile, outputFile)
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

var zgrabDone bool

func launchFullHttpScan(timestampFormatted string, outputPath string, port string, scanTargets string, blacklistFile string) {
	nmapOutputFileName := "zmap_output_" + timestampFormatted + ".csv"

	var cmdZmapString string = "sudo zmap -p " + port + " -n " + scanTargets + " -r " + HTTP_SCAN_DEFAULT_PPS + " -m " + filepath.Join(outputPath, ZMAP_META_DATA_FILE_NAME)

	if blacklistFile != "null" {
		cmdZmapString += " -b " + blacklistFile
	}

	cmdZmapZteeString := cmdZmapString + " | ztee " + filepath.Join(outputPath, nmapOutputFileName)

	metaDataFileName := ZVERSION_META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
	outputFile := filepath.Join(outputPath, getZgrabOutputFilename(timestampFormatted))

	cmdScanString := cmdZmapZteeString + " | zgrab --port " + port + " --data=./http-req --senders 2500 --timeout " + TIMEOUT_IN_SECONDS_FIRST_TRY + " --output-file=" + outputFile + " --metadata-file=" + filepath.Join(outputPath, metaDataFileName)
	content, _ := ioutil.ReadFile("./http-req")
	zgrabRequest = string(content)

	scanCmd := exec.Command("bash", "-c", cmdScanString)

	var wg sync.WaitGroup
	wg.Add(1)
	readerErrOut, writerErrOut := io.Pipe()
	scanCmd.Stderr = writerErrOut

	go progressAndLogZmap(readerErrOut, &wg)

	runErr := scanCmd.Run()
	util.Check(runErr)
	zgrabDone = true

	readerErr1 := readerErrOut.Close()
	util.Check(readerErr1)
	wg.Wait()

	metaDataFile := filepath.Join(outputPath, metaDataFileName)
	enhanceMetaData(metaDataFile, outputFile)
}
func getZgrabOutputFilename(timestampFormatted string) string {
	var zgrabOutputFileName string
	if isVHostScan {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_VHOST + "_" + timestampFormatted + ".json"
	} else {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_FULL + "_" + timestampFormatted + ".json"
	}
	return zgrabOutputFileName
}

//func handleZgrabOutput(currentScanPath string, timestampFormatted string, stdOut io.ReadCloser, wg *sync.WaitGroup) {
//	stdOutScanner := bufio.NewScanner(stdOut)
//
//	zgrabErrorLog := SCAN_ZGRAB_ERROR_FILE_NAME + "_" + timestampFormatted
//	metaDataFileName := ZVERSION_META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
//	metaDataFile, _ := os.Create(filepath.Join(currentScanPath, metaDataFileName))
//	zgrabErr, _ := os.Create(filepath.Join(currentScanPath, zgrabErrorLog))
//	outputFile := filepath.Join(currentScanPath, getZgrabOutputFilename(timestampFormatted))
//	zgrabOut, _ := os.Create(outputFile)
//	successfulFallbackIPs, _ := os.Create(filepath.Join(currentScanPath, ZVERSION_SUCCESSFUL_FALLBACK_IPS))
//
//	writeQueueErr := make(chan []byte, 10000)
//	writeQueueOut := make(chan string, 100000)
//	writeQueueIPs := make(chan string, 10000)
//
//	var wgWriters sync.WaitGroup
//	wgWriters.Add(3)
//	go util.WriteBytesToFile(&wgWriters, writeQueueErr, zgrabErr)
//	go util.WriteStringToFile(&wgWriters, writeQueueOut, zgrabOut)
//	go util.WriteStringToFile(&wgWriters, writeQueueIPs, successfulFallbackIPs)
//
//	var wgWorkers sync.WaitGroup
//	wgWorkers.Add(10)
//	workQueue := make(chan string, 1000000)
//
//	//wait until queue is filled up
//	go func() {
//		time.Sleep(10 * time.Second)
//		//start workers
//		for i := 0; i < 10; i++ {
//			go workOnZgrabOutputLine(workQueue, &wgWorkers, writeQueueErr, writeQueueOut, writeQueueIPs)
//		}
//	}()
//	go periodicFree(5 * time.Minute)
//	for {
//		stdOutScanner.Scan()
//		if stdOutScanner.Text() != "" {
//			workQueue <- stdOutScanner.Text()
//			if cap(workQueue) == len(workQueue) {
//				log.Println("Work queue is full!")
//			}
//		}
//		if zgrabDone {
//			log.Println("Zgrab said done")
//			break
//		}
//	}
//
//	log.Println("reading stdout done")
//	close(workQueue)
//	wgWorkers.Wait()
//
//	close(writeQueueOut)
//	close(writeQueueErr)
//	close(writeQueueIPs)
//
//	wgWriters.Wait()
//	log.Println("Zgrab routines working on lines are done")
//
//	writeMetaData(metaDataString, outputFile, metaDataFile)
//
//	wg.Done()
//
//}
//
//func workOnZgrabOutputLine(workQueue chan string, wg *sync.WaitGroup, writeQueueErr chan []byte, writeQueueOut chan string, writeQueueIPs chan string) {
//	var wgConnections sync.WaitGroup
//	for line := range workQueue {
//		if strings.Contains(line, "success_count") {
//			metaDataString = line
//			continue
//		}
//
//		if strings.Contains(line, `},"error":"`) {
//			wgConnections.Add(1)
//			go handleZgrabError(line, writeQueueOut, writeQueueErr, writeQueueIPs, &wgConnections)
//			//writeQueueOut <- line
//		} else {
//			writeQueueOut <- line
//		}
//
//	}
//	wgConnections.Wait()
//	wg.Done()
//	log.Println("Worker done")
//}
//
//func writeMetaData(line string, outputFile string, metaDateFile *os.File) {
//	defer metaDateFile.Close()
//	var metaData MetaData
//	json.Unmarshal([]byte(line), &metaData)
//
//	timeNow := time.Now()
//	timeThen, err := time.Parse("2006-01-02T15:04:05-07:00", metaData.Start_time)
//	util.Check(err)
//	metaData.End_time = timeNow.Format("2006-01-02T15:04:05-07:00")
//	metaData.Duration = int(timeNow.Sub(timeThen).Seconds())
//	metaData.Timeout_first_try = TIMEOUT_IN_SECONDS_FIRST_TRY_INT
//	metaData.Timeout = 0
//	metaData.Timeout_second_try = TIMEOUT_IN_SECONDS_SECOND_TRY_INT
//	metaData.ZgrabRequest = zgrabRequest
//	metaData.FallbackCount = fallbackCount
//	metaData.Success_count += fallbackCount
//	metaData.Failure_count -= fallbackCount
//	metaData.ScanInputFile = zmapInputFile
//	metaData.ScanOutputFile = outputFile
//	metaData.Sha256OutputFile = util.CalculateSha256(outputFile)
//	j, _ := json.Marshal(metaData)
//
//	w := bufio.NewWriter(metaDateFile)
//
//	w.Write(j)
//	w.WriteString("\n")
//
//	w.Flush()
//
//	os.Stdout.WriteString(string(j) + "\n")
//}

func enhanceMetaData(metaDateFile string, outputFile string) {
	metaDataString, errFile := ioutil.ReadFile(metaDateFile)
	util.Check(errFile)
	var metaData MetaData
	json.Unmarshal([]byte(metaDataString), &metaData)

	metaData.ScanInputFile = zmapInputFile
	metaData.ScanOutputFile = outputFile
	metaData.Sha256OutputFile = util.CalculateSha256(outputFile)
	j, _ := json.Marshal(metaData)
	file, fileErr := os.Open(metaDateFile)
	util.Check(fileErr)
	defer file.Close()
	file.Write(j)
	file.WriteString("\n")

	os.Stdout.WriteString(string(j) + "\n")
}

const IP_REGEX = `((?:\d{1,3}\.){3}\d{1,3})`

//TODO: adjust for vHost scan!
func handleZgrabError(entryString string, outFile chan string, errFile chan []byte, writeQueueIPs chan string, wg *sync.WaitGroup) {
	entry := RawZversionEntry{}
	json.Unmarshal([]byte(entryString), &entry)

	if entry.Error != "" {
		timeout := time.Duration(TIMEOUT_IN_SECONDS_SECOND_TRY_INT * time.Second)
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
				//response.Body.Read(bs)
				meh := bufio.NewReader(response.Body)
				meh.Read(bs)
				if err != nil {
					entry.Error = err.Error()
				} else {
					entry.Body = string(bs)
				}
			}
			if cap(writeQueueIPs) == len(writeQueueIPs) {
				log.Println("Write queue is full!")
			}
			writeQueueIPs <- entry.BaseEntry.IP
			atomic.AddUint32(&fallbackCount, 1)
		} else {
			errFile <- []byte(entry.BaseEntry.IP + ": " + entry.Error)
		}

		j, _ := json.Marshal(entry)
		outFile <- string(j)
	}

	wg.Done()
}

func periodicFree(d time.Duration) {
	tick := time.Tick(d)
	for _ = range tick {
		debug.FreeOSMemory()
	}
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

func progressAndLogZmap(reader io.ReadCloser, wg *sync.WaitGroup) {
	in := bufio.NewScanner(reader)

	for in.Scan() {
		if !strings.Contains(in.Text(), "banner-grab") {
			os.Stderr.WriteString(in.Text() + "\n")
		}
	}
	wg.Done()
}
