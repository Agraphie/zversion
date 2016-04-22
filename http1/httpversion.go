package http1

import (
	"encoding/json"
	"fmt"
	"github.com/agraphie/zversion/worker"
	"sync"
	"strings"
	"unicode"
)

type Entry struct {
	worker.BaseEntry
	Data  struct {
		      Read  string
		      Write string
	      }
	Agent string
	Error string
}

var hosts = struct {
	sync.RWMutex
	m map[string][]Entry
}{m: make(map[string][]Entry)}

func (e Entry) String() string {
	return fmt.Sprintf("IP: %v, Scanned: %v, Agent: %v", e.BaseEntry.IP, e.BaseEntry.Timestamp, e.Agent)
}

func ParseHttpFile(path string) {
	worker.ParseFile(path, workOnLine)
	fmt.Println(hosts.m)
}

func spaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func workOnLine(queue chan string) {
	for line := range queue {
		u := Entry{}
		json.Unmarshal([]byte(line), &u)
		i1 := len(u.Data.Read) > 0
		contains := strings.Contains(u.Data.Read, "Server")

		switch {
		case i1 && contains:
			splittedString := strings.Split(u.Data.Read, "\n")

			for i := range splittedString {
				if strings.Contains(splittedString[i], "Server") {
					serverSplit := strings.Split(splittedString[i], ":")
					u.Agent = spaceMap(serverSplit[1])
					hosts.Lock()
					hosts.m[u.Agent] = append(hosts.m[u.Agent], u)

					hosts.Unlock()
				}
			}
		case i1 && !contains:
			hosts.Lock()
			u.Agent = "Not set"
			hosts.m["No Server"] = append(hosts.m["No Server"], u)
			hosts.Unlock()
		default:
			hosts.Lock()
			hosts.m["Error"] = append(hosts.m["Error"], u)
			hosts.Unlock()
		}
	}
}