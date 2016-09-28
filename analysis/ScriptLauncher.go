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

func RunHTTPAnalyseScriptsOnAllOutputs() {
	defer util.TimeTrack(time.Now(), "Analysing all HTTP files")
	var analysisWaitGroup sync.WaitGroup

	cleanedFiles, _ := ioutil.ReadDir(util.HttpBaseOutputDir)

	for _, f := range cleanedFiles {
		if f.IsDir() {
			httpOutputFolder := filepath.Join(util.HttpBaseOutputDir, f.Name())
			httpOutputFile := filepath.Join(httpOutputFolder, util.HTTP_OUTPUT_FILE_NAME+".json")
			if util.CheckPathExist(httpOutputFile) {
				analysisWaitGroup.Add(1)
				RunHTTPAnalyseScripts(httpOutputFile, httpOutputFolder, &analysisWaitGroup)
			}
		}
	}
	analysisWaitGroup.Wait()
}

func RunSSHAnalyseScriptsOnAllOutputs() {
	defer util.TimeTrack(time.Now(), "Analysing all SSH files")
	var analysisWaitGroup sync.WaitGroup

	cleanedFiles, _ := ioutil.ReadDir(util.SSHBaseOutputDir)

	for _, f := range cleanedFiles {
		if f.IsDir() {
			sshOutputFolder := filepath.Join(util.SSHBaseOutputDir, f.Name())
			sshOutputFile := filepath.Join(sshOutputFolder, util.SSH_OUTPUT_FILE_NAME+".json")
			if util.CheckPathExist(sshOutputFile) {
				analysisWaitGroup.Add(1)
				RunSSHAnalyseScripts(sshOutputFile, sshOutputFolder, &analysisWaitGroup)
			}
		}
	}
	analysisWaitGroup.Wait()
}

func RunHTTPAnalyseScripts(zVersionHTTPOutputFile string, zVersionOutputFolderPath string, folderAnalysisWaitGroup *sync.WaitGroup) {
	httpScriptFolder := filepath.Join(SCRIPT_LIBRARY, SCRIPT_LIBRARY_HTTP)
	launchScripts(httpScriptFolder, zVersionHTTPOutputFile, zVersionOutputFolderPath)
	if folderAnalysisWaitGroup != nil {
		folderAnalysisWaitGroup.Done()
	}
}

func RunSSHAnalyseScripts(zVersionSSHOutputFile string, zVersionOutputFolderPath string, folderAnalysisWaitGroup *sync.WaitGroup) {
	sshScriptFolder := filepath.Join(SCRIPT_LIBRARY, SCRIPT_LIBRARY_SSH)
	launchScripts(sshScriptFolder, zVersionSSHOutputFile, zVersionOutputFolderPath)
	if folderAnalysisWaitGroup != nil {
		folderAnalysisWaitGroup.Done()
	}
}

func launchScripts(scriptFolderPath string, inputFilePath string, outputFolderPath string) {
	defer util.TimeTrack(time.Now(), "Analysing "+inputFilePath)

	files, _ := ioutil.ReadDir(scriptFolderPath)
	scriptOutputFolderPath := filepath.Join(outputFolderPath, SCRIPT_OUTPUT_FOLDER_NAME)
	err := os.MkdirAll(scriptOutputFolderPath, util.FILE_ACCESS_PERMISSION)
	util.Check(err)

	var scriptWaitGroup sync.WaitGroup
	for _, f := range files {
		if !f.IsDir() {
			scriptWaitGroup.Add(1)
			scriptPath := filepath.Join(scriptFolderPath, f.Name())
			launchScript(scriptPath, inputFilePath, scriptOutputFolderPath, &scriptWaitGroup)
		}
	}

	scriptWaitGroup.Wait()
}

func launchScript(scriptPath string, scriptInputFilePath string, scriptOutputFolderPath string, scriptWaitgroup *sync.WaitGroup) {
	fileInfo, err := os.Stat(scriptPath)
	if err == nil {
		//Check if script is executable
		if fileInfo.Mode()&0111 != 0 {
			defer util.TimeTrack(time.Now(), scriptPath)
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
			cmd.Stderr = os.Stderr

			err = cmd.Start()
			if err != nil {
				util.Check(err)
			}
			cmd.Wait()
		} else {
			log.Printf("Script %s is not executable. Skipping.", scriptPath)
		}

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
