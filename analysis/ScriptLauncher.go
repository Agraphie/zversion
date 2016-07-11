package analysis

import (
	"bufio"
	"github.com/agraphie/zversion/util"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const SCRIPT_LIBRARY = "scripts"
const SCRIPT_LIBRARY_HTTP = "http"
const SCRIPT_LIBRARY_SSH = "ssh"
const OUTPUT_FILE_NAME_REGEX = `(?i)(?:Output\sfile\s?name:\s?)(.*)`
const SCRIPT_OUTPUT_FOLDER_NAME = "analysisOutput"

var outputFileNameRegex = regexp.MustCompile(OUTPUT_FILE_NAME_REGEX)

func RunHTTPAnalyseScripts(zVersionHTTPOutputFile string, zVersionOutputFolderPath string) {
	httpScriptFolder := filepath.Join(SCRIPT_LIBRARY, SCRIPT_LIBRARY_HTTP)
	launchScripts(httpScriptFolder, zVersionHTTPOutputFile, zVersionOutputFolderPath)
}

func RunSSHAnalyseScripts(zVersionSSHOutputFile string, zVersionOutputFolderPath string) {
	httpScriptFolder := filepath.Join(SCRIPT_LIBRARY, SCRIPT_LIBRARY_SSH)
	launchScripts(httpScriptFolder, zVersionSSHOutputFile, zVersionOutputFolderPath)
}

func launchScripts(scriptFolderPath string, inputFilePath string, outputFolderPath string) {
	defer util.TimeTrack(time.Now(), "Analysing")

	files, _ := ioutil.ReadDir(scriptFolderPath)
	scriptOutputFolderPath := filepath.Join(outputFolderPath, SCRIPT_OUTPUT_FOLDER_NAME)
	err := os.MkdirAll(scriptOutputFolderPath, util.FILE_ACCESS_PERMISSION)
	util.Check(err)

	var scriptWaitGroup sync.WaitGroup
	for _, f := range files {
		if !f.IsDir() {
			scriptWaitGroup.Add(1)
			scriptPath := filepath.Join(scriptFolderPath, f.Name())
			go launchScript(scriptPath, inputFilePath, scriptOutputFolderPath, &scriptWaitGroup)
		}
	}
	scriptWaitGroup.Wait()
}

func launchScript(scriptPath string, scriptInputFilePath string, scriptOutputFolderPath string, scriptWaitgroup *sync.WaitGroup) {
	if _, err := os.Stat(scriptPath); err == nil {
		cmd := exec.Command(scriptPath, scriptInputFilePath)

		outputFileName := determineOutputFileName(scriptPath)
		if outputFileName == "" {
			scriptFileNameSplit := strings.Split(scriptPath, string(filepath.Separator))
			scriptFileName := scriptFileNameSplit[len(scriptFileNameSplit)-1]
			outputFileName = scriptFileName + ".out"
		}

		// open the out file for writing
		outfile, err := os.Create(filepath.Join(scriptOutputFolderPath, outputFileName))

		if err != nil {
			util.Check(err)
		}
		defer outfile.Close()
		cmd.Stdout = outfile

		err = cmd.Start()
		if err != nil {
			util.Check(err)
		}
		cmd.Wait()
	} else {
		log.Println("Script does not exist! Path: " + scriptPath)
	}

	if scriptWaitgroup != nil {
		scriptWaitgroup.Done()
	}
}

func determineOutputFileName(scriptPath string) string {
	file, err := os.Open(scriptPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var match []string
	var outputFileName string
	for scanner.Scan() {
		match = outputFileNameRegex.FindStringSubmatch(scanner.Text())
		if match != nil {
			outputFileName = match[1]
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return outputFileName
}
