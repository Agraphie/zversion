package main

import (
	"net/http"
	"html/template"
	"os"
	"path/filepath"
	"github.com/agraphie/zversion/http1"
)


var templatesPath = "templates"

func init() {
	dir, _ := os.Getwd() // gives us the source path if we haven't installed.
	templatesPath = filepath.Join(dir, templatesPath)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./html/static"))))

	http.HandleFunc("/", indexViewHandler)
	http.HandleFunc("/httpVersions/", http1.ParseHttpViewHandler)
	//http.HandleFunc("/httpVersions/.*", httpLogViewHandler)

	http.ListenAndServe(":4000", nil)
}



func indexViewHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, nil)
}

