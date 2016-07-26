package worker

import (
	"bufio"
	"github.com/agraphie/zversion/util"
	"log"
	"os"
	"os/exec"
	"regexp"
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
	workQueue := make(chan string, 10000)
	writeQueue := make(chan []byte, 100000)

	// We need to know when everyone is done so we can exit.
	complete := make(chan bool)

	// Read the lines into the work queue.
	go func() {
		var scanner *bufio.Scanner
		isLz4 := regexp.MustCompile(`.*\.lz4`).FindStringSubmatch(inputPath)
		isGz := regexp.MustCompile(`.*\.gz`).FindStringSubmatch(inputPath)

		if isLz4 != nil {
			c1 := exec.Command("lz4", "-dc", inputPath)
			c1.Stderr = os.Stderr
			lz4CatStdout, _ := c1.StdoutPipe()
			scanner = bufio.NewScanner(lz4CatStdout)
			c1.Start()
		} else if isGz != nil {
			c1 := exec.Command("gunzip", "-dc", inputPath)
			c1.Stderr = os.Stderr
			gzCatStdout, _ := c1.StdoutPipe()
			scanner = bufio.NewScanner(gzCatStdout)
			c1.Start()
		} else {
			file, err := os.Open(inputPath)
			if err != nil {
				log.Fatal(err)
			}
			scanner = bufio.NewScanner(file)

			defer file.Close()
		}

		// Close when the functin returns
		//defer file.Close()

		//scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			if len(workQueue) == cap(workQueue) {
				log.Println("Work queue is full! Add more worker routines?")
			}
			workQueue <- scanner.Text()
		}

		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}
		// Close the channel so everyone reading from it knows we're done.
		close(workQueue)
		log.Println("File reader done")
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
	log.Println("Workers done")
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
