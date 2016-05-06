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
	"github.com/agraphie/zversion/http1"

)

var (
	portFlag = flag.String("port", http1.HTTP_SCAN_DEFAULT_PORT, "The port to scan")
	scanTargets = flag.String("targets", "100%", "How many targets should be scanned, absolute or percentage value")
	scanOutputPath = flag.String("scan-output", "scanResults/", "File path to output scan result")
	analysisOutputPath = flag.String("analysis-output", "analysisResults/", "File path to output analysis results")
	isHttpScan = flag.Bool("http-scan", false, "Whether a HTTP scan should be launched")
)

const FILE_ACCESS_PERMISSION = 0755

func init(){
	flag.StringVar(portFlag, "p", "80", "The port to scan")
	flag.StringVar(scanTargets, "t", "100%", "How many targets should be scanned, absolute or percentage value")
	flag.StringVar(scanOutputPath, "so", "scanResults/", "File path to output scan results")
	flag.StringVar(analysisOutputPath, "ao", "analysisResults/", "File path to output analaysis results")
	flag.BoolVar(isHttpScan, "hs", false, "Whether a HTTP scan should be launched")

	flag.Parse()

	if !util.CheckPathExist(*scanOutputPath) {
		err := os.MkdirAll(*scanOutputPath, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
}
func main() {
	if(*isHttpScan){
		fmt.Println("Launching HTTP scan...")
		http1.LaunchHttpScan(nil, *scanOutputPath, *portFlag, *scanTargets)
	} else {
		fmt.Fprintln(os.Stderr, "No scan specified! E.g. specify the flag `-hs` for a complete HTTP scan")
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


