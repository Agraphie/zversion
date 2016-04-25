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
)


var (
 strFlag = flag.String("long-string", "", "Description")
 portFlag = flag.String("port", "80", "The port to scan")
 scanTargets = flag.String("targets", "100%", "How many targets should be scanned, absolute or percentage value")
 scanOutputPath = flag.String("scan-output", "scanResults/", "File path to output scan result")
 analysisOutputPath = flag.String("analysis-output", "analysisResults/", "File path to output analysis results")
)


func init(){
	flag.StringVar(strFlag, "s", "", "Description")
	flag.StringVar(portFlag, "p", "80", "The port to scan")
	flag.StringVar(scanTargets, "n", "100%", "The port to scan")
	flag.StringVar(scanOutputPath, "so", "scanResults/", "File path to output scan results")
	flag.StringVar(analysisOutputPath, "ao", "analysisResults/", "File path to output analaysis results")


	flag.Parse()

}
func main() {
	execCommandWithCancel("sleep 5")
}

func LaunchHttpScan(){

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
func Meh() {
	//sudo zmap -p 80 -n 1000 -q | ztee results.csv | ./zgrab --port 80 --data=./http-req-head --output-file=/home/agraphie/banners5.json
}

