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
)

type httpVersionVars struct {
	Logs    []string
	Banners []string
}

type httpLogVars struct {
	Title   string
	Results HttpVersionResult
}

func httpLogViewHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Split(r.URL.String(), "/")[2]
	file, e := ioutil.ReadFile("http_logs/" + fileName + ".json")

	if e != nil {
		http.Redirect(w, r, "/httpVersions/", http.StatusNotFound)
		return
	}

	var jsonResult HttpVersionResult
	json.Unmarshal(file, &jsonResult)

	httpLogVars := httpLogVars{fileName, jsonResult}
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

		_, err := os.Stat("http_banners/" + fileLocation + ".json")
		_, errWholePath := os.Stat(fileLocation)

		if err == nil {
			ParseHttpFile("http_banners/" + fileLocation + ".json")
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
	files, err := ioutil.ReadDir("http_logs")
	if (err != nil) {
		panic(err)
	}
	var logs []string
	for _, f := range files {
		fileName := strings.Split(f.Name(), ".")[0]
		logs = append(logs, fileName)
	}
	bannerFiles, err := ioutil.ReadDir("http_banners")
	if (err != nil) {
		panic(err)
	}

	var banners []string
	for _, f := range bannerFiles {
		fileName := strings.Split(f.Name(), ".")[0]
		banners = append(banners, fileName)
	}

	return httpVersionVars{logs, banners}
}