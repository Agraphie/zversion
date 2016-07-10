package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/agraphie/zversion/http1"
	"github.com/agraphie/zversion/ssh"
	"github.com/agraphie/zversion/util"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var (
	portFlag           = flag.String("port", http1.HTTP_SCAN_DEFAULT_PORT, "The port to scan")
	scanTargets        = flag.String("targets", "100%", "How many targets should be scanned, absolute or percentage value")
	scanOutputPath     = flag.String("scan-output", "scanResults/", "File path to output scan result")
	analysisOutputPath = flag.String("analysis-output", "analysisResults/", "File path to output analysis results")
	analysisInputPath  = flag.String("analysis-input", "", "Path to zgrab json output")
	blacklistPath      = flag.String("blacklist-file", "", "Path to the blacklist file (has to be in CIDR notation). Type 'null' to launch without blacklist.")
	isHttpScan         = flag.Bool("http-scan", false, "Whether a HTTP scan should be launched")
	isHttpAnalysis     = flag.Bool("http-analysis", false, "Whether a HTTP analysis should be launched")
	isSSHAnalysis      = flag.Bool("ssh-analysis", false, "Whether a SSH analysis should be launched")
)

const FILE_ACCESS_PERMISSION = 0755

func init() {
	flag.StringVar(portFlag, "p", "80", "The port to scan")
	flag.StringVar(scanTargets, "t", "100%", "How many targets should be scanned, absolute or percentage value")
	flag.StringVar(scanOutputPath, "so", "scanResults/", "File path to output scan results")
	flag.StringVar(analysisOutputPath, "ao", "analysisResults/", "File path to output analaysis results")
	flag.StringVar(analysisInputPath, "ai", "", "Path to zgrab json output")
	flag.StringVar(blacklistPath, "bf", "", "Path to the blacklist file (has to be in CIDR notation). Type 'null' to launch without blacklist.")

	flag.BoolVar(isHttpScan, "hs", false, "Whether a HTTP scan should be launched")
	flag.BoolVar(isHttpAnalysis, "ha", false, "Whether a HTTP analysis should be launched")
	flag.BoolVar(isSSHAnalysis, "sa", false, "Whether a SSH analysis should be launched")

	flag.Parse()

	if !util.CheckPathExist(*scanOutputPath) {
		err := os.MkdirAll(*scanOutputPath, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
}
func main() {
	if *isHttpScan {
		if *blacklistPath != "" {
			if *blacklistPath != "null" && !util.CheckPathExist(*blacklistPath) {
				fmt.Fprintln(os.Stderr, "File does not exist or no permission to read it")
			} else {
				fmt.Println("Launching HTTP scan...")
				http1.LaunchHttpScan(nil, *scanOutputPath, *portFlag, *scanTargets, *blacklistPath)
			}
		} else {
			fmt.Fprintln(os.Stderr, "No blacklist file specified! If really scan without blacklist file type '-bf null' as file path")
		}
	} else if *isHttpAnalysis {
		if util.CheckPathExist(*analysisInputPath) {
			log.Printf("Processing file: %s\n", *analysisInputPath)
			http1.ParseHttpFile(*analysisInputPath)
		} else {
			fmt.Printf("File '%s' does not exist or no permission to read it\n", *analysisInputPath)
		}
	} else if *isSSHAnalysis {
		if util.CheckPathExist(*analysisInputPath) {
			log.Printf("Processing file: %s\n", *analysisInputPath)
			ssh.ParseSSHFile(*analysisInputPath)
		} else {
			fmt.Printf("File '%s' does not exist or no permission to read it\n", *analysisInputPath)
		}
	} else {
		fmt.Fprintln(os.Stderr, "No scan or analysis specified! E.g. specify the flag `-hs` for a complete HTTP scan")
	}

}

func execCommandWithCancel(command string) {
	execCommand := strings.Fields(command)
	cmd := exec.Command(execCommand[0], execCommand[1:]...)
	stdout, _ := cmd.StdoutPipe()

	done := make(chan error, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	cmd.Start()

	go func() {
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
