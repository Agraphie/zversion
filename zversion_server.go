package main

import (
	"net/http"
	"html/template"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"fmt"
	"regexp"
	"github.com/agraphie/zversion/http1"
	"encoding/json"
)

type httpVersionVars struct {
	Logs    []string
	Banners []string
}

type httpLogVars struct {
	Title   string
	Results http1.HttpVersionResult
}

var templatesPath = "templates"

func init() {
	dir, _ := os.Getwd() // gives us the source path if we haven't installed.
	templatesPath = filepath.Join(dir, templatesPath)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./html/static"))))

	http.HandleFunc("/", indexViewHandler)
	http.HandleFunc("/httpVersions/", parseHttpViewHandler)
	//http.HandleFunc("/httpVersions/.*", httpLogViewHandler)

	http.ListenAndServe(":4000", nil)
}

func httpLogViewHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Split(r.URL.String(), "/")[2]
	file, e := ioutil.ReadFile("http_logs/" + fileName + ".json")

	if e != nil {
		http.Redirect(w, r, "/httpVersions/", http.StatusNotFound)
		return
	}

	var jsonResult http1.HttpVersionResult
	json.Unmarshal(file, &jsonResult)

	httpLogVars := httpLogVars{fileName, jsonResult}
	t, _ := template.ParseFiles("html/http_log.html")
	t.Execute(w, httpLogVars)
}

func parseHttpViewHandler(w http.ResponseWriter, r *http.Request) {
	match, _ := regexp.MatchString("/httpVersions/http_version_(.*)", r.URL.EscapedPath())

	if r.Method == "GET" && match {
		httpLogViewHandler(w, r)
	} else if r.Method == "POST" {
		fileLocation := r.FormValue("location")
		fmt.Println("checking: " + fileLocation)

		_, err := os.Stat("http_banners/" + fileLocation + ".json")
		_, errWholePath := os.Stat(fileLocation)

		if err == nil {
			http1.ParseHttpFile("http_banners/" + fileLocation + ".json")
			http.Redirect(w, r, "/httpVersions/", http.StatusFound)
			return
		} else if errWholePath == nil {
			http1.ParseHttpFile(fileLocation)
			http.Redirect(w, r, "/httpVersions/", http.StatusFound)
			return
		} else {
			http.Redirect(w, r, "/httpVersions/", http.StatusNotFound)
			return
		}
	} else {
		//	var result map[string][]http1.Entry = http1.ParseHttpFile("/home/agraphie/banners.json")
		httpVersionVars := initialiseHttpVars()
		t, _ := template.ParseFiles("html/http_versions.html")
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

func indexViewHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, nil)
}

