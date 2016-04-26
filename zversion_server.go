package main

import (
	"net/http"
	"os"
	"path/filepath"
	"github.com/agraphie/zversion/controller"
)


var templatesPath = "templates"


func init() {
	dir, _ := os.Getwd() // gives us the source path if we haven't installed.
	templatesPath = filepath.Join(dir, templatesPath)
}

func main() {
	http.HandleFunc(controller.STATIC_URL, controller.StaticHandler)

	http.HandleFunc("/", controller.IndexViewHandler)
	http.HandleFunc("/httpVersions/", controller.ParseHttpViewHandler)
	//http.HandleFunc("/httpVersions/.*", httpLogViewHandler)

	http.ListenAndServe(":4000", nil)
}



