package http1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/util"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HTTP_SCAN_OUTPUTH_PATH = "http/"
const HTTP_SCAN_DEFAULT_PORT = "80"
const HTTP_SCAN_DEFAULT_SCAN_TARGETS = "10000"
const HTTP_SCAN_DEFAULT_PPS = "100000"

type RunningHttpScan struct {
	RunningCommands []*exec.Cmd
	ProgressZmap    float64
	ProgressZgrab   float32
	Started         time.Time
	Finished        time.Time
}

/**
commands is a map where the key is the timestamp when the scan was launched and the values are all cmds which are
running for that timestamp. This makes it easier to kill them off.
*/
func LaunchHttpScan(runningScan *RunningHttpScan, scanOutputPath string, port string, scanTargets string, blacklistFile string) {
	started := time.Now()
	timestampFormatted := started.Format(util.TIMESTAMP_FORMAT)

	if !util.CheckPathExist(scanOutputPath + HTTP_SCAN_OUTPUTH_PATH + timestampFormatted) {
		err := os.MkdirAll(scanOutputPath+HTTP_SCAN_OUTPUTH_PATH+timestampFormatted, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}

	currentScanPath := scanOutputPath + HTTP_SCAN_OUTPUTH_PATH + timestampFormatted + "/"
	nmapOutputFileName := "zmap_output_" + timestampFormatted + ".csv"

	zmapErrorLog := "zmap_error_" + timestampFormatted
	zmapErr, _ := os.Create(currentScanPath + zmapErrorLog)

	defer zmapErr.Close()

	zmapErrW := bufio.NewWriter(zmapErr)

	defer zmapErrW.Flush()

	var c1 *exec.Cmd
	if blacklistFile == "null" {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS)
	} else {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS, "-b", blacklistFile)
	}

	c2 := exec.Command("ztee", currentScanPath+nmapOutputFileName)
	c3 := exec.Command("zgrab", "--port", port, "--data=./http-req-head")
	if runningScan != nil {
		runningScan.RunningCommands = append(runningScan.RunningCommands, c1)
		runningScan.RunningCommands = append(runningScan.RunningCommands, c2)
		runningScan.RunningCommands = append(runningScan.RunningCommands, c3)
		if runningScan.Started.IsZero() {
			runningScan.Started = started
		}
	}

	c1StdErr, _ := c1.StderrPipe()
	c3StdOut, _ := c3.StdoutPipe()
	c3StdErr, _ := c3.StderrPipe()

	c2.Stderr = os.Stderr
	var wg sync.WaitGroup
	wg.Add(1)
	go handleZgrabOutput(currentScanPath, timestampFormatted, c3StdOut, c3StdErr, &wg)
	//go handleZgrabError(currentScanPath+zgrabOutputFileName, c3StdErr, zgrabErrW, c3StdOut)
	//c3.Stderr = zgrabErrW

	c2.Stdin, _ = c1.StdoutPipe()
	c3.Stdin, _ = c2.StdoutPipe()
	//	c3.Stdout = os.Stdout

	_ = c2.Start()
	_ = c3.Start()
	_ = c1.Start()

	if runningScan == nil {
		go printAndLog(c1StdErr, zmapErrW)
	} else {
		go progressAndLogZmap(c1StdErr, zmapErrW, runningScan)
		//go progressZgrab(c1StdErr, c3StdOut, runningScan)
	}

	_ = c2.Wait()
	_ = c3.Wait()
	_ = c1.Wait()

	c1StdErr.Close()
	wg.Wait()

	finished := time.Now()
	if runningScan != nil {
		runningScan.Finished = finished
	}
	log.Printf("Http scan done in: %d ns\n", time.Since(started))
}

func handleZgrabOutput(currentScanPath string, timestampFormatted string, stdOut io.ReadCloser, stdErr io.ReadCloser, wg *sync.WaitGroup) {
	defer wg.Done()
	zgrabOutputFileName := "zgrab_output_" + timestampFormatted + ".json"
	zgrabErrorLog := "zgrab_error_" + timestampFormatted
	zgrabErr, _ := os.Create(currentScanPath + zgrabErrorLog)
	zgrabOut, _ := os.Create(currentScanPath + zgrabOutputFileName)
	defer zgrabErr.Close()
	zgrabErrW := bufio.NewWriter(zgrabErr)
	zgrabOutW := bufio.NewWriter(zgrabOut)
	defer zgrabErrW.Flush()
	defer zgrabOutW.Flush()

	stdOutScanner := bufio.NewScanner(stdOut)
	//zgrabOutputFile, _ := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0600)

	for stdOutScanner.Scan() {
		line := stdOutScanner.Text()
		if strings.Contains(line, "success_count") {
			os.Stdout.WriteString(line + "\n")
			continue
		}

		u := Entry{}
		json.Unmarshal([]byte(line), &u)
		if u.Error != "" {
			handleZgrabError(u, zgrabOutW, zgrabErrW)
		} else {
			zgrabOutW.WriteString(line + "\n")
		}
	}
	stdOut.Close()
	stdErr.Close()
}

func handleZgrabError(entry Entry, outFile *bufio.Writer, errFile *bufio.Writer) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	response, err := client.Get("http://" + entry.BaseEntry.IP)
	if err == nil {
		entry.Agent = response.Header.Get("Server")
		entry.Error = ""
	} else {
		errFile.WriteString(entry.BaseEntry.IP + ": " + entry.Error + "\n")
	}

	j, _ := json.Marshal(entry)
	outFile.WriteString(string(j) + "\n")
}

func printAndLog(reader io.ReadCloser, logWriter io.Writer) {
	in := bufio.NewScanner(reader)

	for in.Scan() {
		logWriter.Write(in.Bytes())
		logWriter.Write([]byte("\n"))

		fmt.Println(in.Text())
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
