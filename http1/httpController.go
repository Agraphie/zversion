package http1

import (
	"strings"
	"io/ioutil"
	"encoding/json"
	"regexp"
	"fmt"
	"os"
	"net/http"
	"html/template"
	"github.com/agraphie/zversion/util"
)

func init(){
	if !util.CheckPathExist(util.SCAN_OUTPUT_BASE_PATH + util.HTTP_SCAN_OUTPUTH_PATH) {
		err := os.MkdirAll(util.SCAN_OUTPUT_BASE_PATH + util.HTTP_SCAN_OUTPUTH_PATH, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
	if !util.CheckPathExist(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH) {
		err := os.MkdirAll(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH, FILE_ACCESS_PERMISSION)
		util.Check(err)
	}
}

type httpVersionVars struct {
	Logs    []string
	Banners []string
}

type httpLogVars struct {
	Title   string
	Results HttpVersionResult
}

func httpLogViewHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Split(r.URL.String(), "/")[2:3]
	file, e := ioutil.ReadFile(util.SCAN_OUTPUT_BASE_PATH + util.HTTP_SCAN_OUTPUTH_PATH + fileName[0] + "/" + fileName[1] + ".json")

	if e != nil {
		http.Redirect(w, r, "/httpVersions/", http.StatusNotFound)
		return
	}

	var jsonResult HttpVersionResult
	json.Unmarshal(file, &jsonResult)

	httpLogVars := httpLogVars{fileName[0] + "/" + fileName[1], jsonResult}
	t, _ := template.ParseFiles("http1/html/http_log.html")
	t.Execute(w, httpLogVars)
}

func ParseHttpViewHandler(w http.ResponseWriter, r *http.Request) {
	match, _ := regexp.MatchString("/httpVersions/http_version_(.*)", r.URL.EscapedPath())

	if r.Method == "GET" && match {
		httpLogViewHandler(w, r)
	} else if r.Method == "POST" {
		fileLocation := r.FormValue("location")
		fmt.Println("checking: " + fileLocation)

		_, err := os.Stat(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH + fileLocation + ".json")
		_, errWholePath := os.Stat(fileLocation)

		if err == nil {
			ParseHttpFile(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH + fileLocation + ".json")
			http.Redirect(w, r, "/httpVersions/", http.StatusFound)
			return
		} else if errWholePath == nil {
			ParseHttpFile(fileLocation)
			http.Redirect(w, r, "/httpVersions/", http.StatusFound)
			return
		} else {
			http.Redirect(w, r, "/httpVersions/", http.StatusNotFound)
			return
		}
	} else {
		httpVersionVars := initialiseHttpVars()
		t, _ := template.ParseFiles("http1/html/http_versions.html")
		t.Execute(w, httpVersionVars)
	}
}

func initialiseHttpVars() httpVersionVars {
	directories, err := ioutil.ReadDir(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH)
	if (err != nil) {
		panic(err)
	}
	var logs []string
	for _, d := range directories {
		files, _ := ioutil.ReadDir(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH + d.Name())

		for _, f := range files{
			fileName := strings.Split(f.Name(), ".")[0]
			logs = append(logs, fileName)
		}

	}
	//bannerFiles, err := ioutil.ReadDir(util.ANALYSIS_OUTPUT_BASE_PATH + util.HTTP_ANALYSIS_OUTPUTH_PATH)
	bannerDirectories, err := ioutil.ReadDir(util.SCAN_OUTPUT_BASE_PATH + util.HTTP_SCAN_OUTPUTH_PATH)

	if (err != nil) {
		panic(err)
	}
	var banners []string

	for _, d := range bannerDirectories {
		files, _ := ioutil.ReadDir(util.SCAN_OUTPUT_BASE_PATH + util.HTTP_SCAN_OUTPUTH_PATH + d.Name())

		for _, f := range files{
			fileName := strings.Split(f.Name(), ".")[0]
			if(strings.Contains(fileName, "zgrab")) {
				banners = append(banners, fileName)
			}
		}

	}


	return httpVersionVars{logs, banners}
}