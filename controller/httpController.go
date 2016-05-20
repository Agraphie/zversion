package controller

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/http1"
	"github.com/agraphie/zversion/util"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const MAPPING = "/httpVersions/"
const SCAN_MAPPING = "/httpVersions/scan"

var runningScans map[string]*http1.RunningHttpScan

func init() {
	runningScans = make(map[string]*http1.RunningHttpScan)
	if !util.CheckPathExist(util.SCAN_OUTPUT_BASE_PATH + http1.HTTP_SCAN_OUTPUTH_PATH) {
		err := os.MkdirAll(util.SCAN_OUTPUT_BASE_PATH+http1.HTTP_SCAN_OUTPUTH_PATH, http1.FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
	if !util.CheckPathExist(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH) {
		err := os.MkdirAll(util.ANALYSIS_OUTPUT_BASE_PATH+util.HTTP_ANALYSIS_OUTPUTH_PATH, http1.FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
}

type httpVersionVars struct {
	Logs         map[string][]string
	Banners      []string
	RunningScans []*http1.RunningHttpScan
}

type httpLogVars struct {
	Title   string
	Results map[string]int
}

func httpLogViewHandler(w http.ResponseWriter, r *http.Request) {
	fileNameSplit := strings.Split(r.URL.String(), "/")
	fileName := fileNameSplit[2] + "/" + fileNameSplit[3]
	file, e := ioutil.ReadFile(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH + fileName + ".json")

	if e != nil {
		http.Redirect(w, r, MAPPING, http.StatusNotFound)
		return
	}

	var jsonResult map[string]int
	json.Unmarshal(file, &jsonResult)
	shortResult := shortenResult(jsonResult)

	httpLogVars := httpLogVars{fileName, shortResult}
	t := render(w, "http_log")
	t.Execute(w, httpLogVars)
}

func shortenResult(results map[string]int) map[string]int{
	newResults := make(map[string]int)

	for k, v := range results {
		if v > 5000{
			newResults[k] = v
		} else {
			newResults["Other"] = newResults["Other"] + v
		}
	}

	return newResults
}

func ParseHttpViewHandler(w http.ResponseWriter, r *http.Request) {
	match, _ := regexp.MatchString(MAPPING+"zgrab_output_(.*)/http_version_(.*)", r.URL.EscapedPath())
	match_scan, _ := regexp.MatchString(SCAN_MAPPING, r.URL.EscapedPath())

	if r.Method == "GET" && match {
		httpLogViewHandler(w, r)
		return
	} else if r.Method == "POST" && match_scan {
		initiateNewHttpScan(w, r)
		return
	} else if r.Method == "POST" {
		fileLocation := r.FormValue("location")
		fmt.Println("checking: " + fileLocation)

		_, err := os.Stat(util.SCAN_OUTPUT_BASE_PATH + http1.HTTP_SCAN_OUTPUTH_PATH + fileLocation + ".json")
		_, errWholePath := os.Stat(fileLocation)

		if err == nil {
			http1.ParseHttpFile(util.SCAN_OUTPUT_BASE_PATH + http1.HTTP_SCAN_OUTPUTH_PATH + fileLocation + ".json")
			http.Redirect(w, r, MAPPING, http.StatusFound)
			return
		} else if errWholePath == nil {
			http1.ParseHttpFile(fileLocation)
			http.Redirect(w, r, MAPPING, http.StatusFound)
			return
		} else {
			log.Println("Path: " + fileLocation + " not found")
			http.NotFound(w, r)
			return
		}
	} else {
		httpVersionVars := initialiseHttpVars()
		t := render(w, "http_versions")
		t.Execute(w, httpVersionVars)
	}
}

func initiateNewHttpScan(w http.ResponseWriter, r *http.Request) {
	newScan := http1.RunningHttpScan{Started: time.Now(), ProgressZgrab: 0, ProgressZmap: 0}
	runningScans[fmt.Sprint(newScan.Started)] = &newScan
	go http1.LaunchHttpScan(&newScan, util.SCAN_OUTPUT_BASE_PATH, http1.HTTP_SCAN_DEFAULT_PORT, http1.HTTP_SCAN_DEFAULT_SCAN_TARGETS, "")

	http.Redirect(w, r, MAPPING, http.StatusFound)
}

func initialiseHttpVars() httpVersionVars {

	analysisLogs := getAnalysisLogs()
	bannerLogs := getBannerLogs()
	scans := getRunningScans()

	return httpVersionVars{Logs: analysisLogs, Banners: bannerLogs, RunningScans: scans}
}

func getRunningScans() []*http1.RunningHttpScan {
	scans := make([]*http1.RunningHttpScan, len(runningScans))
	for _, value := range runningScans {
		scans = append(scans, value)
	}

	return scans
}

func getAnalysisLogs() map[string][]string {
	directories, err := ioutil.ReadDir(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH)
	util.Check(err)
	logs := make(map[string][]string)
	for _, d := range directories {
		files, err := ioutil.ReadDir(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH + d.Name())
		util.Check(err)

		for _, f := range files {
			fileName := strings.Split(f.Name(), ".")[0]
			if strings.Contains(fileName, "summary") {
				logs[d.Name()] = append(logs[d.Name()], fileName)
			}

		}
	}

	return logs
}

func getBannerLogs() []string {
	bannerDirectories, err := ioutil.ReadDir(util.SCAN_OUTPUT_BASE_PATH + http1.HTTP_SCAN_OUTPUTH_PATH)

	if err != nil {
		panic(err)
	}
	var banners []string

	for _, d := range bannerDirectories {
		files, _ := ioutil.ReadDir(util.SCAN_OUTPUT_BASE_PATH + http1.HTTP_SCAN_OUTPUTH_PATH + d.Name())

		for _, f := range files {
			fileName := strings.Split(f.Name(), ".")[0]
			if strings.Contains(fileName, "zgrab_output") {
				banners = append(banners, d.Name()+"/"+fileName)
			}
		}

	}
	return banners
}
