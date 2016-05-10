package controller

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

const STATIC_URL string = "/static/"
const STATIC_ROOT string = "static/"

type Context struct {
	Title  string
	Static string
}

func IndexViewHandler(w http.ResponseWriter, r *http.Request) {
	t := render(w, "index")
	err := t.Execute(w, nil)
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

func render(w http.ResponseWriter, tmpl string) *template.Template {
	tmpl_list := []string{fmt.Sprintf("templates/%s.html", tmpl), "templates/header.html", "templates/footer.html", "templates/navbar.html"}
	t, err := template.ParseFiles(tmpl_list...)
	if err != nil {
		log.Print("template parsing error: ", err)
	}
	return t
}

func StaticHandler(w http.ResponseWriter, req *http.Request) {
	static_file := req.URL.Path[len(STATIC_URL):]
	if len(static_file) != 0 {
		f, err := http.Dir(STATIC_ROOT).Open(static_file)
		if err == nil {
			content := io.ReadSeeker(f)
			http.ServeContent(w, req, static_file, time.Now(), content)
			return
		}
	}
	http.NotFound(w, req)
}
