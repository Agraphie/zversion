package http1

import (
	"os/exec"
	"time"
	"os"
	"bufio"
	"io"
	"fmt"
	"github.com/agraphie/zversion/util"
)
const HTTP_SCAN_OUTPUTH_PATH = "http/"
const HTTP_SCAN_DEFAULT_PORT = "80"
const HTTP_SCAN_DEFAULT_SCAN_TARGETS = "10000"
const HTTP_SCAN_DEFAULT_PPS = "100000"

/**
commands is a map where the key is the timestamp when the scan was launched and the values are all cmds which are
running for that timestamp. This makes it easier to kill them off.
 */
func LaunchHttpScan(commands map[string][]*exec.Cmd, scanOutputPath string, port string, scanTargets string){
	timestamp := time.Now().Format(util.TIMESTAMP_FORMAT)
	if !util.CheckPathExist(scanOutputPath+ HTTP_SCAN_OUTPUTH_PATH+timestamp) {
		err := os.MkdirAll(scanOutputPath+ HTTP_SCAN_OUTPUTH_PATH+timestamp, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}


	currentScanPath := scanOutputPath+HTTP_SCAN_OUTPUTH_PATH+timestamp+"/"
	nmapOutputFileName := "zmap_output_"+timestamp+".csv"
	zgrabOutputFileName := "zgrab_output_" + timestamp + ".json"

	zmapErrorLog := "zmap_error_" + timestamp
	zgrabErrorLog := "zgrab_error_" + timestamp

	zmapErr, _ := os.Create(currentScanPath + zmapErrorLog)
	zgrabErr, _ := os.Create(currentScanPath + zgrabErrorLog)
	defer zmapErr.Close()
	defer zgrabErr.Close()

	zmapErrW := bufio.NewWriter(zmapErr)
	zgrabErrW := bufio.NewWriter(zgrabErr)
	defer zmapErrW.Flush()
	defer zgrabErrW.Flush()

	c1 := exec.Command("sudo", "zmap", "-p", port, "-n", scanTargets, "-r", HTTP_SCAN_DEFAULT_PPS)
	c2 := exec.Command("ztee", currentScanPath+nmapOutputFileName)
	c3 := exec.Command("zgrab", "--port", port, "--data=./http-req-head", "--output-file="+ currentScanPath+zgrabOutputFileName)
	if commands != nil{
		commands[timestamp] = append(commands[timestamp], c1)
		commands[timestamp] = append(commands[timestamp], c2)
		commands[timestamp] = append(commands[timestamp], c3)
	}

	c1StdErr, _ := c1.StderrPipe()
	c2.Stderr = os.Stderr
	c3.Stderr = zgrabErrW

	c2.Stdin, _ = c1.StdoutPipe()
	c3.Stdin, _ = c2.StdoutPipe()
	c3.Stdout = os.Stdout

	_ = c2.Start()
	_ = c3.Start()
	_ = c1.Start()

	go printAndLog(c1StdErr, zmapErrW)

	_ = c2.Wait()
	_ = c3.Wait()
	_ = c1.Wait()
}

func printAndLog(reader io.ReadCloser, logWriter io.Writer){
	in := bufio.NewScanner(reader)

	for in.Scan() {
		logWriter.Write(in.Bytes())
		logWriter.Write([]byte("\n"))

		fmt.Println(in.Text())
	}
}