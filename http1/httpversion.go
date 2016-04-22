package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/worker"
)

type Entry struct {
	worker.BaseEntry
	Data  struct {
		      Read  string
		      Write string
	      }
	Agent string
}

func startWorking(queue chan string) {
	for line := range queue {
		u := &Entry{}
		json.Unmarshal([]byte(line), &u)
		fmt.Println(u.Data.Read)
	}
}

func ParseHttpFile(path string) {
	httpEntry := func(queue chan string) {
		for line := range queue {
			u := &Entry{}
			json.Unmarshal([]byte(line), &u)
			fmt.Println(u)
		}
	}

	worker.ParseFile(path, httpEntry)
}