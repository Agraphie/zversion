package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/agraphie/zversion/analysis"
	"github.com/agraphie/zversion/http1"
	"github.com/agraphie/zversion/ssh"
	"github.com/agraphie/zversion/util"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	portFlag       = flag.String("port", http1.HTTP_SCAN_DEFAULT_PORT, "The port to scan")
	scanTargets    = flag.String("targets", "100%", "How many targets should be scanned, absolute or percentage value")
	scanOutputPath = flag.String("scan-output", "scanResults/", "File path to output scan result")
	scanInputFile  = flag.String("scan-input", "", "An input file containing one IP or URL per line which will be scanned. If none specified, a full scan will be launched.")

	analysisOutputPath = flag.String("analysis-output", "analysisResults/", "File path to output analysis results")
	analysisInputPath  = flag.String("analysis-input", "", "Path to zgrab json output")
	blacklistPath      = flag.String("blacklist-file", "", "Path to the blacklist file (has to be in CIDR notation). Type 'null' to launch without blacklist.")
	isHttpScan         = flag.Bool("http-scan", false, "Whether a HTTP scan should be launched")
	isHttpAnalysis     = flag.Bool("http-analysis", false, "Whether a HTTP analysis should be launched")
	isSSHAnalysis      = flag.Bool("ssh-analysis", false, "Whether a SSH analysis should be launched")
	rerunScripts       = flag.String("run-scripts", "", "Rerun all scripts on target or all cleaned files")
)

const FILE_ACCESS_PERMISSION = 0755
const RERUN_SCRIPTS_ON_ALL_CLEANED_FILES_FLAG = "all"

func init() {
	flag.StringVar(portFlag, "p", "80", "The port to scan")
	flag.StringVar(scanTargets, "t", "100%", "How many targets should be scanned, absolute or percentage value")
	flag.StringVar(scanOutputPath, "so", "scanResults/", "File path to output scan results")
	flag.StringVar(scanInputFile, "si", "", "An input file containing one IP or URL per line which will be scanned. If none specified, a full scan will be launched.")

	flag.StringVar(analysisOutputPath, "ao", "analysisResults/", "File path to output analaysis results")
	flag.StringVar(analysisInputPath, "ai", "", "Path to zgrab json output")
	flag.StringVar(blacklistPath, "bf", "", "Path to the blacklist file (has to be in CIDR notation). Type 'null' to launch without blacklist.")
	flag.StringVar(rerunScripts, "rs", "", "Path to target cleaned directory on which the scripts should be run. Can only be used in conjunction with 'http-analysis' or 'ssh-analysis'. Entering Type 'all' to run all scripts on all cleaned files.")

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
		if *blacklistPath != "" || *scanInputFile != "" {
			if (*blacklistPath != "null" && !util.CheckPathExist(*blacklistPath)) && *scanInputFile == "" {
				fmt.Fprintf(os.Stderr, "File '%s' does not exist or no permission to read it\n", *blacklistPath)
			} else if *scanInputFile != "" && !util.CheckPathExist(*scanInputFile) {
				fmt.Fprintf(os.Stderr, "File '%s' does not exist or no permission to read it\n", *scanInputFile)
			} else {
				fmt.Println("Launching HTTP scan...")
				http1.LaunchHttpScan(nil, *scanOutputPath, *portFlag, *scanTargets, *blacklistPath, *scanInputFile)
			}
		} else {
			fmt.Fprintln(os.Stderr, "No blacklist file specified! If really scan without blacklist file type '-bf null' as file path")
		}
	} else if *isHttpAnalysis && *rerunScripts == "" {
		if util.CheckPathExist(*analysisInputPath) {
			log.Printf("Processing file: %s\n", *analysisInputPath)
			http1.ParseHttpFile(*analysisInputPath)
		} else {
			fmt.Printf("File '%s' does not exist or no permission to read it\n", *analysisInputPath)
		}
	} else if *isSSHAnalysis && *rerunScripts == "" {
		if util.CheckPathExist(*analysisInputPath) {
			log.Printf("Processing file: %s\n", *analysisInputPath)
			ssh.ParseSSHFile(*analysisInputPath)
		} else {
			fmt.Printf("File '%s' does not exist or no permission to read it\n", *analysisInputPath)
		}
	} else if (*isHttpAnalysis || *isSSHAnalysis) && *rerunScripts != "" {
		if *rerunScripts == RERUN_SCRIPTS_ON_ALL_CLEANED_FILES_FLAG {
			log.Println("Starting analysis of all output files")

			if *isSSHAnalysis {
				analysis.RunSSHAnalyseScriptsOnAllOutputs()
			} else {
				analysis.RunHTTPAnalyseScriptsOnAllOutputs()
			}
		} else if util.CheckPathExist(*rerunScripts) {
			log.Printf("Analysing file output in folder: %s\n", *rerunScripts)
			if *isSSHAnalysis {
				analysis.RunSSHAnalyseScripts(filepath.Join(*rerunScripts, ssh.OUTPUT_FILE_NAME+".json"), *rerunScripts, nil)
			} else {
				analysis.RunHTTPAnalyseScripts(filepath.Join(*rerunScripts, util.HTTP_OUTPUT_FILE_NAME+".json"), *rerunScripts, nil)
			}
		} else {
			fmt.Printf("File '%s' does not exist or no permission to read it\n", *rerunScripts)
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
