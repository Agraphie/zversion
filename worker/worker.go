package worker

import (
	"bufio"
	"github.com/agraphie/zversion/util"
	"log"
	"os"
	"sync"
	"time"
)

type BaseEntry struct {
	IP        string
	Timestamp time.Time
	Error     string
}

type HostsConcurrentSafe struct {
	sync.RWMutex
	M map[string]int
}

func ParseFile(inputPath string, outputFile *os.File, f func(queue chan string, complete chan bool, hosts *HostsConcurrentSafe, writeQueue chan []byte)) HostsConcurrentSafe {
	var hosts = HostsConcurrentSafe{M: make(map[string]int)}
	// This channel has no buffer, so it only accepts input when something is ready
	// to take it out. This keeps the reading from getting ahead of the writers.
	workQueue := make(chan string)
	writeQueue := make(chan []byte)

	// We need to know when everyone is done so we can exit.
	complete := make(chan bool)

	// Read the lines into the work queue.
	go func() {
		file, err := os.Open(inputPath)
		if err != nil {
			log.Fatal(err)
		}

		// Close when the functin returns
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			workQueue <- scanner.Text()
		}

		// Close the channel so everyone reading from it knows we're done.
		close(workQueue)
	}()

	//start writer
	go util.WriteEntries(complete, writeQueue, outputFile)

	// Now read them all off, concurrently.
	for i := 0; i < util.CONCURRENCY; i++ {
		go f(workQueue, complete, &hosts, writeQueue)
	}

	// Wait for everyone to finish.
	for i := 0; i < util.CONCURRENCY; i++ {
		<-complete
	}

	close(writeQueue)

	//wait for write queue
	<-complete

	return hosts
}

func AddToMap(key string, hosts *HostsConcurrentSafe) {
	hosts.Lock()
	hosts.M["Total Processed"] = hosts.M["Total Processed"] + 1
	hosts.M[key] = hosts.M[key] + 1
	hosts.Unlock()
}
