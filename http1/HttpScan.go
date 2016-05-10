package http1

import (
	"bufio"
	"fmt"
	"github.com/agraphie/zversion/util"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	zgrabOutputFileName := "zgrab_output_" + timestampFormatted + ".json"

	zmapErrorLog := "zmap_error_" + timestampFormatted
	zgrabErrorLog := "zgrab_error_" + timestampFormatted

	zmapErr, _ := os.Create(currentScanPath + zmapErrorLog)
	zgrabErr, _ := os.Create(currentScanPath + zgrabErrorLog)
	defer zmapErr.Close()
	defer zgrabErr.Close()

	zmapErrW := bufio.NewWriter(zmapErr)
	zgrabErrW := bufio.NewWriter(zgrabErr)
	defer zmapErrW.Flush()
	defer zgrabErrW.Flush()

	var c1 *exec.Cmd
	if blacklistFile == "null" {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS)
	} else {
		c1 = exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS, "-b", blacklistFile)
	}

	c2 := exec.Command("ztee", currentScanPath+nmapOutputFileName)
	c3 := exec.Command("zgrab", "--port", port, "--data=./http-req-head", "--output-file="+currentScanPath+zgrabOutputFileName)
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

	c2.Stderr = os.Stderr
	c3.Stderr = zgrabErrW

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
	c3StdOut.Close()
	finished := time.Now()
	if runningScan != nil {
		runningScan.Finished = finished
	}
	log.Printf("Http scan done in: %d ns\n", time.Since(started))
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
