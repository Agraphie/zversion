package worker

import (
	"bufio"
	"log"
	"os"
	"time"
)

type BaseEntry struct {
	IP        string
	Timestamp time.Time
	Error     string
}

var concurrency = 1

func ParseFile(inputPath string, f func(queue chan string)) {
	// This channel has no buffer, so it only accepts input when something is ready
	// to take it out. This keeps the reading from getting ahead of the writers.
	workQueue := make(chan string)

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

	// Now read them all off, concurrently.
	for i := 0; i < concurrency; i++ {
		go startWorking(workQueue, complete, f)
	}

	// Wait for everyone to finish.
	for i := 0; i < concurrency; i++ {
		<-complete
	}
}

func startWorking(queue chan string, complete chan bool, f func(queue chan string)) {
	f(queue)

	// Let the main process know we're done.
	complete <- true
}
