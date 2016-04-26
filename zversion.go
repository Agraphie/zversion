package main

import (
	"flag"
	"os/exec"
	"log"
	"strings"
	"fmt"
	"bufio"
	"os"
	"os/signal"
	"syscall"
	"errors"
	"github.com/agraphie/zversion/util"
	"time"
	"io"
)

var (
	portFlag = flag.String("port", "80", "The port to scan")
	scanTargets = flag.String("targets", "10000", "How many targets should be scanned, absolute or percentage value")
	scanOutputPath = flag.String("scan-output", "scanResults/", "File path to output scan result")
	analysisOutputPath = flag.String("analysis-output", "analysisResults/", "File path to output analysis results")
	httpScan = flag.Bool("http-scan", false, "Wether a HTTP scan should be launched")
)

const FILE_ACCESS_PERMISSION = 0755

func init(){
	flag.StringVar(portFlag, "p", "80", "The port to scan")
	flag.StringVar(scanTargets, "n", "10000", "The port to scan")
	flag.StringVar(scanOutputPath, "so", "scanResults/", "File path to output scan results")
	flag.StringVar(analysisOutputPath, "ao", "analysisResults/", "File path to output analaysis results")
	flag.BoolVar(httpScan, "hs", false, "Wether a HTTP scan should be launched")

	flag.Parse()

	if !util.CheckPathExist(*scanOutputPath) {
		err := os.MkdirAll(*scanOutputPath, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
}
func main() {
	if(*httpScan){
		fmt.Println("Launching HTTP scan...")
		LaunchHttpScan(nil)
	}

}
/**
commands is a map where the key is the timestamp when the scan was launched and the values are all cmds which are
running for that timestamp. This makes it easier to kill them off.
 */
func LaunchHttpScan(commands map[string][]*exec.Cmd){
	timestamp := time.Now().Format(util.TIMESTAMP_FORMAT)
	if !util.CheckPathExist(*scanOutputPath+util.HTTP_SCAN_OUTPUTH_PATH+timestamp) {
		err := os.MkdirAll(*scanOutputPath+util.HTTP_SCAN_OUTPUTH_PATH+timestamp, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}


	currentScanPath := *scanOutputPath+util.HTTP_SCAN_OUTPUTH_PATH+timestamp+"/"
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

	c1 := exec.Command("sudo", "zmap", "-p", *portFlag, "-n", *scanTargets, "-r 100000")
	c2 := exec.Command("ztee", currentScanPath+nmapOutputFileName)
	c3 := exec.Command("zgrab", "--port", *portFlag, "--data=./http-req-head", "--output-file="+ currentScanPath+zgrabOutputFileName)
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

func execCommandWithCancel(command string){
	execCommand := strings.Fields(command)
	cmd := exec.Command(execCommand[0], execCommand[1:]...)
	stdout, _ := cmd.StdoutPipe()


	done := make(chan error, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	cmd.Start()

	go func(){
		sig := <-sigs
		switch sig {
		case os.Interrupt:
			done <- errors.New("Interrupted")
		case syscall.SIGTERM:
			done <- errors.New("Terminated")
		}
	}()

	go func() {
		done <- cmd.Wait()
	}()





	in := bufio.NewScanner(stdout)

	for in.Scan() {
		fmt.Println("meh")
	}

	select {
	case err := <-done:
		if err != nil {
			//cmd.Process.Kill()
			log.Printf("process done with error = %v", err)
		} else {
			log.Print("process done gracefully without error")
		}
	}
}


