package ssh

import (
	"bufio"
	"encoding/json"
	"github.com/agraphie/zversion/util"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	SSH_SCAN_OUTPUTH_PATH          = "ssh"
	SSH_SCAN_DEFAULT_PORT          = "22"
	HTTP_SCAN_DEFAULT_SCAN_TARGETS = "10000"
	SSH_SCAN_DEFAULT_PPS           = "100000"
	SCAN_OUTPUT_FILE_NAME_FULL     = "zversion_ssh_full"
	SCAN_OUTPUT_FILE_NAME_VHOST    = "zversion_ssh_vhost"
	SCAN_ZGRAB_ERROR_FILE_NAME     = "zgrab_error"
	ZVERSION_META_DATA_FILE_NAME   = "zversion_scan_meta_data"
	ZMAP_META_DATA_FILE_NAME       = "zmap_scan_meta_data.json"
)

type RunningHttpScan struct {
	RunningCommands []*exec.Cmd
	ProgressZmap    float64
	ProgressZgrab   float32
	Started         time.Time
	Finished        time.Time
}

type RawZversionEntry struct {
	BaseEntry
	Data struct {
		Read string `json:",omitempty"`
	} `json:",omitempty"`
	Error string
	Body  string
}

type MetaData struct {
	Port             int     `json:"port"`
	Success_count    uint32  `json:"success_count"`
	Failure_count    uint32  `json:"failure_count"`
	Total            int     `json:"total"`
	Start_time       string  `json:"start_time"`
	End_time         string  `json:"end_time"`
	Duration         int     `json:"duration"`
	Senders          int     `json:"senders"`
	Timeout          int     `json:"timeout"`
	Tls_version      *string `json:"tls_version"`
	Mail_type        *string `json:"mail_type"`
	Ca_file_name     *string `json:"ca_file_name"`
	Sni_support      bool    `json:"sni_support"`
	ScanInputFile    *string `json:"input_file"`
	ScanOutputFile   string  `json:"output_file"`
	Sha256OutputFile string  `json:"sha256_sum_of_output_file"`
}

var zmapInputFile *string
var isVHostScan bool

const TIMEOUT_IN_SECONDS = "5"
const TIMEOUT_IN_SECONDS_INT = 5
const MAX_KB_TO_READ = "64"

/**
commands is a map where the key is the timestamp when the scan was launched and the values are all cmds which are
running for that timestamp. This makes it easier to kill them off.
*/
const FILE_ACCESS_PERMISSION = 0755

func LaunchSSHScan(scanOutputPath string, port string, scanTargets string, blacklistFile string, inputFile string) {
	started := time.Now()
	timestampFormatted := started.Format(util.TIMESTAMP_FORMAT_SECONDS)

	outputPath := filepath.Join(scanOutputPath, SSH_SCAN_OUTPUTH_PATH, timestampFormatted)
	if !util.CheckPathExist(outputPath) {
		err := os.MkdirAll(outputPath, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}

	if inputFile == "" {
		zmapInputFile = nil
		launchFullHttpScan(timestampFormatted, outputPath, port, scanTargets, blacklistFile)
	} else {
		zmapInputFile = &inputFile
		isVHostScan = checkVHostScan(inputFile)
		launchRestrictedHttpScan(outputPath, timestampFormatted, port, inputFile)
	}
	log.Printf("SSH scan done in: %s\n", time.Since(started))
}

func launchRestrictedHttpScan(outputPath string, timestampFormatted string, port string, inputFile string) {
	var zgrabOutputFileName string
	if isVHostScan {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_VHOST + "_" + timestampFormatted + ".json"
	} else {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_FULL + "_" + timestampFormatted + ".json"
	}
	outputFile := filepath.Join(outputPath, zgrabOutputFileName)
	metaDataFileName := ZVERSION_META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
	metaDataFile := filepath.Join(outputPath, metaDataFileName)

	cmdScanString := "zgrab --ssh --port " + port + " --timeout " + TIMEOUT_IN_SECONDS + " --output-file=" + outputFile + " --metadata-file=" + metaDataFile + " --input-file=" + inputFile

	scanCmd := exec.Command("bash", "-c", cmdScanString)
	var wg sync.WaitGroup
	wg.Add(1)
	readerErrOut, writerErrOut := io.Pipe()
	scanCmd.Stderr = writerErrOut

	go progressAndLogZmap(readerErrOut, &wg)
	runErr := scanCmd.Run()
	util.Check(runErr)

	readerErr1 := readerErrOut.Close()
	util.Check(readerErr1)
	wg.Wait()
	enhanceMetaData(metaDataFile, outputFile)
}

func checkVHostScan(inputFile string) bool {
	file, err := os.Open(inputFile)
	util.Check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	firstLine := scanner.Text()

	util.Check(scanner.Err())

	splitComma := strings.Split(firstLine, ",")

	return len(splitComma) == 2
}

func launchFullHttpScan(timestampFormatted string, outputPath string, port string, scanTargets string, blacklistFile string) {
	nmapOutputFileName := "zmap_output_" + timestampFormatted + ".csv"

	var zgrabOutputFileName string
	if isVHostScan {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_VHOST + "_" + timestampFormatted + ".json"
	} else {
		zgrabOutputFileName = SCAN_OUTPUT_FILE_NAME_FULL + "_" + timestampFormatted + ".json"
	}
	outputFile := filepath.Join(outputPath, zgrabOutputFileName)

	var cmdZmapString string = "sudo zmap -p " + port + " -n " + scanTargets + " -r " + SSH_SCAN_DEFAULT_PPS + " -m " + filepath.Join(outputPath, ZMAP_META_DATA_FILE_NAME)

	if blacklistFile != "null" {
		cmdZmapString += " -b " + blacklistFile
	}

	cmdZmapZteeString := cmdZmapString + " | ztee " + filepath.Join(outputPath, nmapOutputFileName)

	metaDataFileName := ZVERSION_META_DATA_FILE_NAME + "_" + timestampFormatted + ".json"
	metaDataFile := filepath.Join(outputPath, metaDataFileName)

	cmdScanString := cmdZmapZteeString + " | zgrab --ssh --port " + port + " --timeout " + TIMEOUT_IN_SECONDS + " --output-file=" + outputFile + " --metadata-file=" + metaDataFile

	scanCmd := exec.Command("bash", "-c", cmdScanString)

	var wg sync.WaitGroup
	wg.Add(1)
	readerErrOut, writerErrOut := io.Pipe()
	scanCmd.Stderr = writerErrOut

	go progressAndLogZmap(readerErrOut, &wg)
	runErr := scanCmd.Run()
	util.Check(runErr)

	readerErr1 := readerErrOut.Close()
	util.Check(readerErr1)
	wg.Wait()
	enhanceMetaData(metaDataFile, outputFile)
}

func enhanceMetaData(metaDateFile string, outputFile string) {
	metaDataString, errFile := ioutil.ReadFile(metaDateFile)
	util.Check(errFile)
	var metaData MetaData
	json.Unmarshal([]byte(metaDataString), &metaData)
	metaData.ScanInputFile = zmapInputFile
	metaData.ScanOutputFile = outputFile
	metaData.Sha256OutputFile = util.CalculateSha256(outputFile)
	j, _ := json.Marshal(metaData)
	file, fileErr := os.Open(metaDateFile)
	util.Check(fileErr)
	defer file.Close()
	file.Write(j)
	file.WriteString("\n")

	os.Stdout.WriteString(string(j) + "\n")
}

func progressAndLogZmap(reader io.ReadCloser, wg *sync.WaitGroup) {
	in := bufio.NewScanner(reader)

	for in.Scan() {
		if !strings.Contains(in.Text(), "banner-grab") {
			os.Stderr.WriteString(in.Text() + "\n")
		}
	}
	wg.Done()
}
